package skin

import (
	"context"
	"os"
	"path/filepath"
	"sort"

	"github.com/zellydev-games/opensplit/logger"
)

const logModule = "skins"

// Service allows for platform switching of skins
type Service struct {
	ctx     context.Context
	skinDir string
}

// Startup takes a context.Context passed by Wails.run OnStartup and sets it to this Service.
func (s *Service) Startup(ctx context.Context, skinDir string) {
	s.ctx = ctx
	s.skinDir = skinDir
}

// GetAvailableSkins walks the skins folder and reports the folders that have a valid skin structure
func (s *Service) GetAvailableSkins() []string {
	var availableSkins []string
	entries, err := os.ReadDir(s.skinDir)
	if err != nil {
		logger.Errorf(logModule, "failed to read skins directory: %s", err.Error())
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
