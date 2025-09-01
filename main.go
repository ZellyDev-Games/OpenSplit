package main

import (
	"OpenSplit/logger"
	"OpenSplit/session"
	"OpenSplit/session/persister"
	"OpenSplit/timer"
	"context"
	"embed"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"github.com/google/uuid"
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

	timerService, timeUpdatedChannel := timer.NewService()
	logger.Debug("Timer service initialized")

	jsonFilePersister := persister.JsonFile{}
	logger.Debug("JSON FilePersister initialized")

	mockSegments := []session.Segment{
		*session.NewSegment(uuid.New(), "Segment 1"),
		*session.NewSegment(uuid.New(), "Segment 2"),
		*session.NewSegment(uuid.New(), "Segment 3"),
	}
	mockSplitFile := session.NewSplitFile("Test Game", "Any%", mockSegments)
	sessionService := session.NewService(timerService, timeUpdatedChannel, jsonFilePersister, mockSplitFile)

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
			sessionService.Startup(ctx)
			runtime.WindowSetAlwaysOnTop(ctx, true)
		},
		Bind: []interface{}{
			sessionService,
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
