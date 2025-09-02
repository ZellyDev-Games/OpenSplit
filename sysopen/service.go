package sysopen

import (
	"OpenSplit/logger"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type Service struct {
	ctx        context.Context
	skinFolder string
}

func NewService(skinFolder string) *Service {
	return &Service{
		skinFolder: skinFolder,
	}
}

func (s *Service) Startup(ctx context.Context) {
	s.ctx = ctx
}

func (s *Service) OpenSkinsFolder() {
	err := s.OpenFolder(s.skinFolder)
	if err != nil {
		logger.Error(err.Error())
	}
}

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
