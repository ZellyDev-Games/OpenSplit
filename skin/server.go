package skin

import (
	"fmt"
	"net"
	"net/http"
	"path/filepath"

	"github.com/zellydev-games/opensplit/logger"
)

// InitListener creates the local HTTP server used to serve skin assets to the frontend.
// The listener is initialized only once.
func (s *Service) InitListener() (int, error) {
	s.initOnce.Do(func() {
		fs := http.FileServer(http.Dir(s.skinDir))

		mux := http.NewServeMux()

		mux.Handle("/skin/", http.StripPrefix("/skin/", http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			// Disable browser caching.
			//
			// This is required for:
			// - live skin editing
			// - CSS reloads after file changes
			// - preview updates
			w.Header().Set(
				"Cache-Control",
				"no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0",
			)

			w.Header().Set(
				"Pragma",
				"no-cache",
			)

			w.Header().Set(
				"Expires",
				"0",
			)

			w.Header().Set(
				"Surrogate-Control",
				"no-store",
			)

			if filepath.Ext(r.URL.Path) == ".css" {
				w.Header().Set(
					"Content-Type",
					"text/css; charset=utf-8",
				)
			}

			fs.ServeHTTP(w, r)
		})))

		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			s.initErr = err
			logger.Errorf(
				logModule,
				"error creating listener: %s",
				err.Error(),
			)
			return
		}

		s.listener = ln

		addr := ln.Addr().(*net.TCPAddr)

		s.port = addr.Port

		s.address = fmt.Sprintf(
			"http://127.0.0.1:%d/skin/",
			s.port,
		)

		s.server = &http.Server{
			Handler: mux,
		}

		logger.Infof(
			logModule,
			"skin server setup on %s",
			s.address,
		)
	})

	return s.port, s.initErr
}

// Serve starts the skin HTTP server in a background goroutine.
func (s *Service) Serve() {
	logger.Info(logModule, "starting skin server")
	s.serving.Store(true)
	go func() {
		err := s.server.Serve(s.listener)
		if err != nil {
			logger.Errorf(logModule, "error serving skin server: %s", err.Error())
		}

		s.serving.Store(false)
		logger.Info(logModule, "skin server stopped")
	}()
}

// Serving reports whether the embedded HTTP server is currently running.
func (s *Service) Serving() bool {
	return s.serving.Load()
}

// Addr returns the base URL of the local skin server.
func (s *Service) Addr() string {
	return s.address
}
