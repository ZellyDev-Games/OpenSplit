package main

import (
	"OpenSplit/hotkeys"
	"OpenSplit/logger"
	"OpenSplit/session"
	"OpenSplit/skin"
	"OpenSplit/sysopen"
	"OpenSplit/timer"
	"context"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

var (
	shutdownOnce sync.Once
	shutdownDone = make(chan struct{})
)

func main() {
	_, logDir, skinDir := setupPaths()

	setupLogging(logDir)
	logger.Info("logging initialized, starting opensplit")

	sysopenService := sysopen.NewService(skinDir)
	logger.Debug("sysopen service initialized")

	skinsFileServer := setupSkinServer(skinDir)
	skinService := &skin.Service{}
	logger.Debug("Skin service initialized")

	timerService, timeUpdatedChannel := timer.NewService()
	logger.Debug("Timer service initialized")

	jsonFilePersister := session.JsonFile{}
	logger.Debug("JSON FilePersister initialized")

	sessionService := session.NewService(timerService, timeUpdatedChannel, nil, jsonFilePersister)
	logger.Debug("SessionService initialized")

	hotkeyProvider, keyInfoChannel := setupHotkeys()
	hotkeyService := hotkeys.NewService(keyInfoChannel, sessionService, hotkeyProvider)
	logger.Debug("HotkeyService initialized")

	logger.Info("services initialized, starting application")
	err := wails.Run(&options.App{
		Title:     "OpenSplit",
		Width:     1024,
		Height:    768,
		Frameless: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if len(r.URL.Path) > 7 && r.URL.Path[:7] == "/skins/" {
					skinsFileServer.ServeHTTP(w, r)
					return
				}
				http.NotFound(w, r)
			}),
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			sysopenService.Startup(ctx)
			timerService.Startup(ctx)
			skinService.Startup(ctx, skinDir)
			sessionService.Startup(ctx)
			wailsRuntime.WindowSetAlwaysOnTop(ctx, true)
			err := hotkeyProvider.StartHook()
			if err != nil {
				logger.Error(fmt.Sprintf("failed to start hotkey provider hook: %s", err))
				os.Exit(2)
			}
			hotkeyService.StartDispatcher()
			logger.Info("application startup complete")

			go func() {
				ch := make(chan os.Signal, 1)
				signal.Notify(ch, os.Interrupt, syscall.SIGTERM) // disables default exit for these
				s := <-ch
				logger.Info(fmt.Sprintf("received exit signal %s", s))

				// Do cleanup *now* so we don't depend on Wails calling OnShutdown
				gracefulShutdown(hotkeyProvider)

				// Ask Wails to quit (this will still call OnShutdown in normal paths)
				wailsRuntime.Quit(ctx)

				// Give Wails a brief moment to unwind; then hard-exit if needed
				select {
				case <-shutdownDone:
				case <-time.After(2 * time.Second):
				}
				os.Exit(0)
			}()
		},
		OnShutdown: func(ctx context.Context) {
			gracefulShutdown(hotkeyProvider)
		},
		Bind: []interface{}{
			sessionService,
			sysopenService,
			skinService,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func setupSkinServer(skinDir string) http.Handler {
	return http.StripPrefix("/skins/", http.FileServer(http.Dir(skinDir)))
}

func setupPaths() (string, string, string) {
	base, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	appDir := filepath.Join(base, "OpenSplit")
	logDir := filepath.Join(appDir, "logs")
	skinDir := filepath.Join(appDir, "skins")
	err = os.MkdirAll(appDir, 755)
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(logDir, 755)
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(skinDir, 755)
	if err != nil {
		panic(err)
	}
	return appDir, logDir, skinDir
}

func setupLogging(logDir string) {
	f, err := os.OpenFile(path.Join(logDir, "OpenSplit.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
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

func gracefulShutdown(hotkeyProvider hotkeys.HotkeyProvider) {
	shutdownOnce.Do(func() {
		_ = hotkeyProvider.Unhook()
		logger.Info("shutdown complete")
		close(shutdownDone)
	})
}

func setupHotkeys() (hotkeys.HotkeyProvider, chan hotkeys.KeyInfo) {
	var hotkeyProvider hotkeys.HotkeyProvider
	var keyInfoChannel chan hotkeys.KeyInfo
	switch runtime.GOOS {
	case "windows":
		hotkeyProvider, keyInfoChannel = hotkeys.NewWindowsHotkeyManager()
	default:
		logger.Error(fmt.Sprintf("operating system %s is not yet supported", runtime.GOOS))
		os.Exit(1)
	}

	return hotkeyProvider, keyInfoChannel
}
