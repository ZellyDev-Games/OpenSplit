package skin

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zellydev-games/opensplit/config"
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/repo"
)

//go:embed default-skin.zip
var DefaultSkinZip []byte

const logModule = "skins"
const EntryPoint = "index.css"

type DirectoryWatcher interface {
	Start(string, func())
	ChangeRoot(string)
}

// Service allows for platform switching of skins
type Service struct {
	m             sync.RWMutex
	initOnce      sync.Once
	initErr       error
	skinDir       string
	server        *http.Server
	listener      net.Listener
	serving       atomic.Bool
	address       string
	port          int
	selectedSkin  string
	configService *config.Service
	repoService   *repo.Service
	watcher       DirectoryWatcher
	skinUpdatedCh chan string
}

func NewService(skinDir string,
	config *config.Service,
	repoService *repo.Service,
	watcher DirectoryWatcher) (*Service, chan string) {
	ch := make(chan string)
	return &Service{
		skinDir:       skinDir,
		configService: config,
		repoService:   repoService,
		watcher:       watcher,
		skinUpdatedCh: ch,
	}, ch
}

// Startup takes a context.Context passed by Wails.run OnStartup and sets it to this Service.
func (s *Service) Startup() error {
	if s.skinDir == "" {
		return errors.New("skinDir not set")
	}

	if !filepath.IsAbs(s.skinDir) {
		msg := fmt.Sprintf("skinDir must be absolute: %s", s.skinDir)
		logger.Error(logModule, msg)
		return errors.New(msg)
	}

	target := s.skinDir
	if !strings.Contains(target, "OpenSplit") {
		msg := fmt.Sprintf("refusing to delete outside OpenSplit directory: %s", target)
		logger.Error(logModule, msg)
		return errors.New(msg)
	}

	if err := os.RemoveAll(filepath.Join(target, "default")); err != nil {
		return err
	}

	err := os.MkdirAll(target, 0o755)
	if err != nil {
		return err
	}

	r, err := zip.NewReader(bytes.NewReader(DefaultSkinZip), int64(len(DefaultSkinZip)))
	if err != nil {
		return err
	}

	for _, f := range r.File {
		p := filepath.Join(target, f.Name)

		// Prevent ZipSlip
		if !strings.HasPrefix(filepath.Clean(p)+string(os.PathSeparator), filepath.Clean(target)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal zip path: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(p, 0o755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		out, err := os.OpenFile(p, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			_ = rc.Close()
			return err
		}

		_, err = io.Copy(out, rc)
		_ = out.Close()
		_ = rc.Close()
		if err != nil {
			return err
		}
	}

	if s.configService.SelectedSkin == "" {
		err := s.SetSkin("default", true)
		if err != nil {
			// we've exhausted all our options, just manually set to default
			// so that the thing at least runs.
			logger.Warnf(logModule, "failed to set default skin, manual fallback: %v", err)
			s.m.Lock()
			s.selectedSkin = "default"
			s.m.Unlock()
		}
	} else {
		err := s.SetSkin(s.configService.SelectedSkin, false)
		if err != nil {
			logger.Warnf(logModule,
				"failed to set selected skin %s, fallback setskin to default: %v",
				s.configService.SelectedSkin, err)
			err = s.SetSkin("default", true)
			if err != nil {
				logger.Warnf(logModule, "default setskin fallback failed, manual fallback: %v", err)
				// we've exhausted all our options, just manually set to default
				// so that the thing at least runs.
				s.m.Lock()
				s.selectedSkin = "default"
				s.m.Unlock()
			}
		}
	}

	s.m.RLock()
	watchDir := filepath.Join(s.skinDir, s.selectedSkin)
	s.m.RUnlock()

	s.watcher.Start(watchDir, s.skinUpdated)
	return nil
}

// GetAvailableSkins walks the skins folder and reports the folders that have a valid skin structure
func (s *Service) GetAvailableSkins() []string {
	var availableSkins []string
	s.m.RLock()
	defer s.m.RUnlock()

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
		cssPath := filepath.Join(s.skinDir, name, EntryPoint)
		if info, err := os.Stat(cssPath); err == nil && !info.IsDir() {
			availableSkins = append(availableSkins, name)
		}
	}
	sort.Strings(availableSkins)
	return availableSkins
}

func (s *Service) SetSkin(name string, writeConfig bool) error {
	for _, skin := range s.GetAvailableSkins() {
		if name == skin {
			s.m.Lock()
			s.selectedSkin = name
			s.m.Unlock()

			s.watcher.ChangeRoot(filepath.Join(s.skinDir, s.selectedSkin))

			if writeConfig {
				s.configService.SelectedSkin = name
				err := s.repoService.SaveConfig(s.configService)
				if err != nil {
					logger.Errorf(logModule, "failed to save config: %s", err.Error())
					return err
				}
			}
			return nil
		}
	}

	logger.Errorf(logModule, "skin %s not found", name)
	return errors.New("skin not found")
}

func (s *Service) SelectedSkin() string {
	s.m.RLock()
	defer s.m.RUnlock()
	return s.selectedSkin
}

func (s *Service) GetSkinAddress() string {
	s.m.RLock()
	defer s.m.RUnlock()

	u, _ := url.Parse(s.address)
	u.Path = path.Join(u.Path, s.selectedSkin, EntryPoint)
	return u.String()
}

func (s *Service) skinUpdated() {
	s.skinUpdatedCh <- s.GetSkinAddress() + "?v=" + strconv.FormatInt(time.Now().UnixNano(), 10)
}
