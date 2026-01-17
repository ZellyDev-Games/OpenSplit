package repo

import (
	"errors"
	"sync"

	"github.com/zellydev-games/opensplit/config"
	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/repo/adapters"
	"github.com/zellydev-games/opensplit/session"
)

// ErrConfigMissing signals to the caller that the config file is not there (first run, or user moved it), so generate a default
var ErrConfigMissing = errors.New("config missing")

// Repository defines a contract for a repo provider to operate against
type Repository interface {
	LoadSplitFile() ([]byte, error)
	GetLoadedSplitFile() ([]byte, error)
	SaveSplitFile([]byte, string) error
	SaveAs([]byte, string) error
	ClearCachedFileName()
	SaveConfig([]byte) error
	LoadConfig() ([]byte, error)
}

type Service struct {
	splitFileLock sync.RWMutex
	configLock    sync.RWMutex
	repository    Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

// LoadSplitFile reads splitfile bytes from a repo and returns it as a session.SplitFile
func (s *Service) LoadSplitFile() (session.SplitFile, error) {
	s.splitFileLock.RLock()
	splitFile, err := s.repository.LoadSplitFile()
	if err != nil {
		s.splitFileLock.RUnlock()
		return session.SplitFile{}, err
	}
	s.splitFileLock.RUnlock()
	splitFileSTO, _ := adapters.JSONSplitFileToDTO(string(splitFile))
	return adapters.DTOSplitFileToDomain(splitFileSTO)
}

// SaveSplitFileWindowDimensions loads the active filename in the repository service,
// modified the window dimension fields in that file, and resaves it without touching split or run data
func (s *Service) SaveSplitFileWindowDimensions(X int, Y int, Width int, Height int) error {
	s.splitFileLock.RLock()
	diskSplitFileBytes, err := s.repository.GetLoadedSplitFile()
	if err != nil {
		s.splitFileLock.RUnlock()
		return err
	}
	s.splitFileLock.RUnlock()

	diskSplitFile, err := adapters.JSONSplitFileToDTO(string(diskSplitFileBytes))
	if err != nil {
		return err
	}

	diskSplitFile.WindowX = X
	diskSplitFile.WindowY = Y
	diskSplitFile.WindowWidth = Width
	diskSplitFile.WindowHeight = Height

	return s.SaveSplitFile(diskSplitFile)
}

func (s *Service) SaveSplitFile(splitFile dto.SplitFile) error {
	payload, err := adapters.SplitFileToFrontEnd(splitFile)
	if err != nil {
		return err
	}
	identifier := splitFile.GameName
	if splitFile.GameCategory != "" {
		identifier += "-" + splitFile.GameCategory
	}
	identifier += ".osf"

	// minimum sizes and position
	splitFile.WindowX = max(10, splitFile.WindowX)
	splitFile.WindowY = max(10, splitFile.WindowY)
	splitFile.WindowWidth = max(100, splitFile.WindowWidth)
	splitFile.WindowHeight = max(100, splitFile.WindowHeight)

	s.splitFileLock.Lock()
	defer s.splitFileLock.Unlock()
	return s.repository.SaveSplitFile(payload, identifier)
}

func (s *Service) Close() {
	s.splitFileLock.Lock()
	defer s.splitFileLock.Unlock()
	s.repository.ClearCachedFileName()
}

func (s *Service) SaveConfig(configService *config.Service) error {
	payload, err := adapters.ConfigToFrontEnd(configService)
	if err != nil {
		return err
	}
	s.configLock.Lock()
	defer s.configLock.Unlock()
	return s.repository.SaveConfig(payload)
}

func (s *Service) LoadConfig(c *config.Service) error {
	s.configLock.RLock()
	b, err := s.repository.LoadConfig()
	if err != nil {
		s.configLock.RUnlock()
		return err
	}
	s.configLock.RUnlock()

	newConfig, err := adapters.FrontEndToConfig(b)
	if err != nil {
		return err
	}

	s.configLock.Lock()
	c.SpeedRunAPIBase = newConfig.SpeedRunAPIBase
	c.KeyConfig = newConfig.KeyConfig
	s.configLock.Unlock()
	return nil
}
