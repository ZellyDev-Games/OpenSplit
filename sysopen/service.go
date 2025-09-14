package sysopen

import (
	"context"
	"errors"
	"github.com/zellydev-games/opensplit/logger"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Service provides a binding to allow the frontend to open a folder in the native OS file explorer.
type Service struct {
	ctx        context.Context
	skinFolder string
}

// NewService returns a Service that can be used to open a folder in the native OS file explorer.
func NewService(skinFolder string) *Service {
	return &Service{
		skinFolder: skinFolder,
	}
}

// Startup passes in a context from Wails.Run to allow the exec call to open a file explorer.
//
// Any other context will cause OpenFolder and OpenSkinsFolder to panic.
func (s *Service) Startup(ctx context.Context) {
	s.ctx = ctx
}

// OpenSkinsFolder opens the skins folder in the native OS file explorer.
func (s *Service) OpenSkinsFolder() {
	err := s.OpenFolder(s.skinFolder)
	if err != nil {
		logger.Error(err.Error())
	}
}

// OpenFolder opens an arbitrary file path in the native OS file explorer.
func (s *Service) OpenFolder(path string) error {
	if path == "" {
		return errors.New("empty path")
	}
	abs := path
	if !filepath.IsAbs(path) {
		a, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		abs = a
	}
	info, err := os.Stat(abs)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		logger.Warn("sysopen.OpenFolder not called with directory path")
		return nil
	}

	switch runtime.GOOS {
	case "windows":
		return exec.CommandContext(s.ctx, "explorer", filepath.FromSlash(abs)).Start()
	case "darwin":
		return exec.CommandContext(s.ctx, "open", abs).Start()
	case "linux":
		return exec.CommandContext(s.ctx, "xdg-open", abs).Start()
	default:
		return errors.New("unsupported OS")
	}
}
