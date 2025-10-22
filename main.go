package main

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/zellydev-games/opensplit/autosplitters"
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

var (
	shutdownOnce sync.Once
	shutdownDone = make(chan struct{})
)

func main() {
	_, logDir, _ := setupPaths()
	setupLogging(logDir)
	logger.Info("logging initialized, starting opensplit")

	runtimeProvider := platform.NewWailsRuntime()
	fileProvider := platform.NewFileRuntime()
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

	// NWA example
	// // This should try to re/connect if used
	// // {
	autosplitterService := autosplitters.NewService(true, true, 48879, "0.0.0.0", autosplitters.NWA)
	// // time.Sleep(1 * time.Second)
	// // }

	autosplitterService.EmuInfo()
	autosplitterService.EmuGameInfo()
	autosplitterService.EmuStatus()
	autosplitterService.ClientID()
	autosplitterService.CoreInfo()
	autosplitterService.CoreMemories()

	// //this is the core loop of autosplitting
	// //queries the device (emu, hardware, application) at the rate specified in ms
	// {
	autoState, err2 := autosplitterService.Update()
	if err2 != nil {
		return
	}
	if autoState.Start {
		//start run
	}
	if autoState.Reset {
		//restart run
	}
	if autoState.Split {
		//split run
	}
	// time.Sleep(time.Duration(1000.0/pollingRate) * time.Millisecond)
	// }

	// //QUSB2SNES example
	// client, err := ConnectSyncClient()
	// client.SetName("annelid")

	// version, err := client.AppVersion()
	// fmt.Printf("Server version is %v\n", version)

	// devices, err := client.ListDevice()

	// if len(devices) != 1 {
	// 	if len(devices) == 0 {
	// 		return errors.New("no devices present")
	// 	}
	// 	return errors.Errorf("unexpected devices: %#v", devices)
	// }
	// device := devices[0]
	// fmt.Printf("Using device %v\n", device)

	// client.Attach(device)
	// fmt.Println("Connected.")

	// info, err := client.Info()
	// fmt.Printf("%#v\n", info)

	// var autosplitter AutoSplitter = NewSuperMetroidAutoSplitter(settings)

	// for {
	// 	summary, err := autosplitter.Update(client)
	// 	if summary.Start {
	// 		timer.Start()
	// 	}
	// 	if summary.Reset {
	// 		if resetTimerOnGameReset == true {
	// 			timer.Reset(true)
	// 		}
	// 	}
	// 	if summary.Split {
	// 		// IGT
	// 		timer.SetGameTime(*t)
	// 		// RTA
	// 		timer.Split()
	// 	}

	// 	if ev == TimerReset {
	// 		// creates a new SNES state
	// 		autosplitter.ResetGameTracking()
	// 		if resetGameOnTimerReset == true {
	// 			client.Reset()
	// 		}
	// 	}

	// 	time.Sleep(time.Duration(float64(time.Second) / pollingRate))
	// }

	// Build dispatcher that can receive commands from frontend or backend and dispatch them to the state machine
	commandDispatcher := dispatcher.NewService(machine)

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

			startInterruptListener(ctx, hotkeyProvider)
			runtime.WindowSetAlwaysOnTop(ctx, true)
			logger.Info("application startup complete")
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
		println("Error:", err.Error())
	}
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

//func setupSkinServer(skinDir string) http.Handler {
//	return http.StripPrefix("/skins/", http.FileServer(http.Dir(skinDir)))
//}

func setupPaths() (string, string, string) {
	base, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	appDir := filepath.Join(base, "OpenSplit")
	logDir := filepath.Join(appDir, "logs")
	skinDir := filepath.Join(appDir, "skins")
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
	return appDir, logDir, skinDir
}

func startInterruptListener(ctx context.Context, hotkeyProvider statemachine.HotkeyProvider) {
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM) // disables default exit for these
		s := <-ch
		logger.Info(fmt.Sprintf("received exit signal %s", s))

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
		logger.Info("shutdown complete")
		close(shutdownDone)
	})
}
