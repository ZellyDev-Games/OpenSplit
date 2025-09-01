package main

import (
	timer2 "OpenSplit2/hotkeys/timer"
	"OpenSplit2/logger"
	"OpenSplit2/timer"
	"context"
	"embed"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	_, logDir := setupPaths()
	setupLogging(logDir)
	logger.Info("logging initialized, starting opensplit")

	timerService := timer.NewService()
	timerHotkeys := timer2.TimerHotkeys{}
	// Create application with options
	err := wails.Run(&options.App{
		Title:  "OpenSplit",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			timerService.Startup(ctx)
			timerHotkeys.Startup(ctx, timerService)
			runtime.WindowSetAlwaysOnTop(ctx, true)
		},
		Bind: []interface{}{
			timerService,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func setupPaths() (string, string) {
	base, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	appDir := filepath.Join(base, "OpenSplit")
	return appDir, fmt.Sprintf("%s/logs", appDir)
}

func setupLogging(logDir string) {
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		log.Fatalln("Failed to create app and log directory:", err)
	}

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
