package main

import (
	"OpenSplit/logger"
	"OpenSplit/persister"
	"OpenSplit/session"
	"OpenSplit/skin"
	"OpenSplit/sysopen"
	"OpenSplit/timer"
	"context"
	"embed"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	_, logDir, skinDir := setupPaths()

	setupLogging(logDir)
	logger.Info("logging initialized, starting opensplit")

	sysopenService := sysopen.NewService(skinDir)

	skinsFileServer := setupSkinServer(skinDir)
	logger.Info("skin server initialized")

	skinService := &skin.Service{}

	timerService, timeUpdatedChannel := timer.NewService()
	logger.Debug("Timer service initialized")

	jsonFilePersister := persister.JsonFile{}
	persisterService := persister.NewService(&jsonFilePersister)
	logger.Debug("JSON FilePersister initialized")

	sessionService := session.NewService(timerService, timeUpdatedChannel, nil)
	logger.Debug("SessionService initialized")

	// Create application with options
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
			persisterService.Startup(ctx)
			skinService.Startup(ctx, skinDir)
			sessionService.Startup(ctx)
			//runtime.WindowSetAlwaysOnTop(ctx, true)
			logger.Info("startup complete")
		},
		Bind: []interface{}{
			sessionService,
			persisterService,
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
