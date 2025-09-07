package skin

import (
	"OpenSplit/logger"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type Service struct {
	ctx     context.Context
	skinDir string
}

func (s *Service) Startup(ctx context.Context, skinDir string) {
	s.ctx = ctx
	s.skinDir = skinDir
}

func (s *Service) GetAvailableSkins() []string {
	var availableSkins []string
	entries, err := os.ReadDir(s.skinDir)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to read skins directory: %s", err.Error()))
		return []string{}
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		cssPath := filepath.Join(s.skinDir, name, name+".css")
		if info, err := os.Stat(cssPath); err == nil && !info.IsDir() {
			availableSkins = append(availableSkins, name)
		}
	}
	sort.Strings(availableSkins)
	return availableSkins
}
