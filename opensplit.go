package main

import (
	"compress/gzip"
	"context"
	"embed"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/zellydev-games/opensplit/autosplitters"
	nwa "github.com/zellydev-games/opensplit/autosplitters/NWA"
	qusb2snes "github.com/zellydev-games/opensplit/autosplitters/QUSB2SNES"
	"github.com/zellydev-games/opensplit/bridge"
	"github.com/zellydev-games/opensplit/config"
	"github.com/zellydev-games/opensplit/dispatcher"
	"github.com/zellydev-games/opensplit/hotkeys"
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/platform"
	"github.com/zellydev-games/opensplit/repo"
	"github.com/zellydev-games/opensplit/session"
	"github.com/zellydev-games/opensplit/statemachine"
	"github.com/zellydev-games/opensplit/timer"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

const logModule = "main"

var (
	shutdownOnce sync.Once
	shutdownDone = make(chan struct{})
)

func main() {
	runtimeProvider := platform.NewWailsRuntime()
	fileProvider := platform.NewFileRuntime()

	_, logDir, _, _, _ := setupPaths(fileProvider)
	setupLogging(logDir)
	logger.Info(logModule, "logging initialized, starting opensplit")

	jsonRepo := repo.NewJsonFile(runtimeProvider, fileProvider)

	timerService, timerUpdateChannel := timer.NewStopwatch(timer.NewTicker(time.Millisecond * 20))
	repoService := repo.NewService(jsonRepo)
	configService, configUpdateChannel := config.NewService()

	sessionService, sessionUpdateChannel := session.NewService(timerService)
	machine := statemachine.InitMachine(runtimeProvider, repoService, sessionService, configService)

	// Build UI bridges with model update channels
	timerUIBridge := bridge.NewTimer(timerUpdateChannel, runtimeProvider)
	sessionUIBridge := bridge.NewSession(sessionUpdateChannel, runtimeProvider)
	configUIBridge := bridge.NewConfig(configUpdateChannel, runtimeProvider)

	// Build dispatcher that can receive commands from frontend or backend and dispatch them to the state machine
	commandDispatcher := dispatcher.NewService(machine)

	// UseAutoSplitter and Type should come from the splits config file
	// ResetTimerOnGameReset, ResetGameOnTimerReset, Addr, Port should come from the autosplitter config file
	// NWA
	AutoSplitterService := autosplitters.Splitters{
		NWAAutoSplitter:       new(nwa.NWASplitter),
		QUSB2SNESAutoSplitter: new(qusb2snes.SyncClient),
		UseAutosplitter:       true,
		ResetTimerOnGameReset: true,
		ResetGameOnTimerReset: false,
		Addr:                  "0.0.0.0",
		Port:                  48879,
		Type:                  autosplitters.NWA}

	// // QUSB2SNES
	// AutoSplitterService := autosplitters.Splitters{
	// 	NWAAutoSplitter:       new(nwa.NWASplitter),
	// 	QUSB2SNESAutoSplitter: new(qusb2snes.SyncClient),
	// 	UseAutosplitter:       true,
	// 	ResetTimerOnGameReset: true,
	// 	ResetGameOnTimerReset: false,
	// 	Addr:                  "0.0.0.0",
	// 	Port:                  23074,
	// 	Type:                  autosplitters.QUSB2SNES}

	var hotkeyProvider statemachine.HotkeyProvider

	err := wails.Run(&options.App{
		Title:     "OpenSplit",
		Width:     1024,
		Height:    768,
		Frameless: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if len(r.URL.Path) > 7 && r.URL.Path[:7] == "/skins/" {
					//skinsFileServer.ServeHTTP(w, r)
					return
				}
				http.NotFound(w, r)
			}),
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			hotkeyProvider = hotkeys.SetupHotkeys()
			machine.AttachHotkeyProvider(hotkeyProvider)
			timerService.Startup(ctx)
			runtimeProvider.Startup(ctx)
			machine.Startup(ctx)

			// Start UI pumps
			sessionUIBridge.StartUIPump()
			timerUIBridge.StartUIPump()
			configUIBridge.StartUIPump()

			//Start autosplitter
			AutoSplitterService.Run(commandDispatcher)

			startInterruptListener(ctx, hotkeyProvider)
			runtime.WindowSetAlwaysOnTop(ctx, true)
			runtime.WindowSetMinSize(ctx, 100, 100)
			logger.Info(logModule, "application startup complete")
		},
		OnBeforeClose: func(ctx context.Context) bool {
			gracefulShutdown(hotkeyProvider)
			return false
		},
		Bind: []interface{}{
			commandDispatcher,
		},
	})

	if err != nil {
		logger.Error(logModule, err.Error())
		os.Exit(1)
	}
}

func setupLogging(logDir string) {
	logPath := path.Join(logDir, "OpenSplit.log")

	// Rotate + compress existing log file, if present
	if _, err := os.Stat(logPath); err == nil {
		ts := time.Now().Format("20060102-150405")
		rotated := logPath + "." + ts
		compressed := rotated + ".gz"

		if err := os.Rename(logPath, rotated); err != nil {
			panic(err)
		}

		if err := compressFile(rotated, compressed); err != nil {
			panic(err)
		}

		_ = os.Remove(rotated)
	}

	// Enforce retention limit
	enforceLogRetention(logDir, "OpenSplit.log", 10)

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}

	consoleHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	fileHandler := slog.NewJSONHandler(f, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger.AddHandler(consoleHandler)
	logger.AddHandler(fileHandler)
}

//func setupSkinServer(skinDir string) http.Handler {
//	return http.StripPrefix("/skins/", http.FileServer(http.Dir(skinDir)))
//}

func setupPaths(fileProvider repo.FileProvider) (string, string, string, string, string) {
	base, err := fileProvider.UserHomeDir()
	if err != nil {
		panic(err)
	}

	appDir := filepath.Join(base, "OpenSplit")
	logDir := filepath.Join(appDir, "logs")
	skinDir := filepath.Join(appDir, "Skins")
	autosplittersDir := filepath.Join(appDir, "Autosplitters")
	splitFileDir := filepath.Join(appDir, "Split Files")
	err = os.MkdirAll(appDir, 0755)
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(logDir, 0755)
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(skinDir, 0755)
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(splitFileDir, 0755)
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(autosplittersDir, 0755)
	if err != nil {
		panic(err)
	}
	return appDir, logDir, skinDir, splitFileDir, autosplittersDir
}

func startInterruptListener(ctx context.Context, hotkeyProvider statemachine.HotkeyProvider) {
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM) // disables default exit for these
		s := <-ch
		logger.Infof(logModule, "received exit signal %s", s)

		// Do cleanup *now* so we don't depend on Wails calling OnShutdown
		if hotkeyProvider != nil {
			gracefulShutdown(hotkeyProvider)
		}

		// Ask Wails to quit (this will still call OnShutdown in normal paths)
		runtime.Quit(ctx)

		// Give Wails a brief moment to unwind; then hard-exit if needed
		select {
		case <-shutdownDone:
		case <-time.After(2 * time.Second):
		}
		os.Exit(0)
	}()
}

func gracefulShutdown(hotkeyService statemachine.HotkeyProvider) {
	shutdownOnce.Do(func() {
		_ = hotkeyService.Unhook()
		logger.Info(logModule, "shutdown complete")
		close(shutdownDone)
	})
}

func compressFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(in *os.File) {
		_ = in.Close()
	}(in)

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		_ = out.Close()
	}(out)

	gz := gzip.NewWriter(out)
	defer func(gz *gzip.Writer) {
		_ = gz.Close()
	}(gz)

	_, err = io.Copy(gz, in)
	return err
}

func enforceLogRetention(dir, base string, max int) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	type logFile struct {
		name string
		mod  time.Time
	}

	var logs []logFile

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, base+".") && strings.HasSuffix(name, ".gz") {
			info, err := e.Info()
			if err != nil {
				continue
			}
			logs = append(logs, logFile{
				name: name,
				mod:  info.ModTime(),
			})
		}
	}

	if len(logs) <= max {
		return
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].mod.Before(logs[j].mod)
	})

	for i := 0; i < len(logs)-max; i++ {
		_ = os.Remove(path.Join(dir, logs[i].name))
	}
}
