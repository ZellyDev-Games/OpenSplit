package repo

import (
	"errors"
	"strings"
	"sync"

	"github.com/zellydev-games/opensplit/config"
	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/repo/adapters"
	"github.com/zellydev-games/opensplit/session"
)

const logModule = "repo"

// ErrConfigMissing signals to the caller that the config file is not there (first run, or user moved it), so generate a default
var ErrConfigMissing = errors.New("config missing")

// Repository defines a contract for a repo provider to operate against
type Repository interface {
	LoadSplitFile() ([]byte, error)
	GetLoadedSplitFile() ([]byte, error)
	SaveSplitFile([]byte, string, bool) error
	SaveAs([]byte, string) error
	Export([]byte, string) error
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

func buildSplitFileName(sf dto.SplitFile) string {
	var parts []string

	parts = append(parts, sf.Platform)
	parts = append(parts, sf.GameName)
	parts = append(parts, sf.GameCategory)

	for _, v := range sf.Variables {
		if strings.TrimSpace(v.Label) != "" {
			parts = append(parts, v.Label)
		}
	}

	return strings.Join(parts, "-") + ".osf"
}

func (s *Service) SaveSplitFile(splitFile dto.SplitFile) error {
	// Merge statistics from the currently loaded split file when
	// saving a newer split file version.
	s.splitFileLock.RLock()
	existingBytes, err := s.repository.GetLoadedSplitFile()
	s.splitFileLock.RUnlock()

	if err == nil {
		existingDTO, err := adapters.JSONSplitFileToDTO(string(existingBytes))
		if err == nil {
			existingDomain, err := adapters.DTOSplitFileToDomain(existingDTO)
			if err == nil {
				newDomain, err := adapters.DTOSplitFileToDomain(splitFile)
				if err == nil {
					if newDomain.Version > existingDomain.Version {
						logger.Infof(
							logModule,
							"upgrading split file v%d -> v%d",
							existingDomain.Version,
							newDomain.Version,
						)

						session.UpgradeSplitFile(&existingDomain, &newDomain)

						splitFile = adapters.DomainSplitFileToDTO(newDomain)
					}
				}
			}
		}
	}

	// minimum sizes and position
	splitFile.WindowX = max(10, splitFile.WindowX)
	splitFile.WindowY = max(10, splitFile.WindowY)
	splitFile.WindowWidth = max(100, splitFile.WindowWidth)
	splitFile.WindowHeight = max(100, splitFile.WindowHeight)

	payload, err := adapters.SplitFileToFrontEnd(splitFile)
	if err != nil {
		return err
	}
	identifier := buildSplitFileName(splitFile)

	logger.Debugf(logModule, "repository saving split file: %s", identifier)
	s.splitFileLock.Lock()
	err = s.repository.SaveSplitFile(payload, identifier, false)
	s.splitFileLock.Unlock()
	if err != nil {
		logger.Errorf(logModule, "repo failed to save splitfile: %s", err)
		return err
	}

	logger.Infof(logModule, "repository saved split file: %s", identifier)
	return nil
}

// LoadSplitFile reads splitfile bytes from a repo and returns it as a session.SplitFile
func (s *Service) LoadSplitFile() (dto.SplitFile, error) {
	logger.Debug(logModule, "loading split file")
	s.splitFileLock.RLock()
	splitFile, err := s.repository.LoadSplitFile()
	if err != nil {
		s.splitFileLock.RUnlock()
		return dto.SplitFile{}, err
	}
	s.splitFileLock.RUnlock()
	splitFileDTO, _ := adapters.JSONSplitFileToDTO(string(splitFile))
	logger.Infof(logModule, "loaded split file: %s-%s", splitFileDTO.GameName, splitFileDTO.GameCategory)
	return splitFileDTO, nil
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

	logger.Debugf(logModule, "saving window dimensions: X: %d, Y: %d, Width: %d, Height: %d", X, Y, Width, Height)
	return s.SaveSplitFile(diskSplitFile)
}

func (s *Service) SaveWorldRecordDisplay(show bool) error {
	s.splitFileLock.Lock()
	defer s.splitFileLock.Unlock()

	existingBytes, err := s.repository.GetLoadedSplitFile()
	if err != nil {
		return err
	}

	dto, err := adapters.JSONSplitFileToDTO(string(existingBytes))
	if err != nil {
		return err
	}

	dto.WR.Show = show

	payload, err := adapters.SplitFileToFrontEnd(dto)
	if err != nil {
		return err
	}

	identifier := buildSplitFileName(dto)

	return s.repository.SaveSplitFile(
		payload,
		identifier,
		false,
	)
}

// SaveSplitFileLayout updates only the persisted splitter layout.
func (s *Service) SaveSplitFileLayout(layout string) error {
	if layout != "vertical" && layout != "horizontal" {
		return errors.New("invalid splitter layout")
	}

	s.splitFileLock.Lock()
	defer s.splitFileLock.Unlock()

	existingBytes, err := s.repository.GetLoadedSplitFile()
	if err != nil {
		return err
	}

	dto, err := adapters.JSONSplitFileToDTO(string(existingBytes))
	if err != nil {
		return err
	}

	dto.Layout = layout

	payload, err := adapters.SplitFileToFrontEnd(dto)
	if err != nil {
		return err
	}

	identifier := buildSplitFileName(dto)

	logger.Debugf(
		logModule,
		"saving splitter layout: %s",
		layout,
	)

	return s.repository.SaveSplitFile(
		payload,
		identifier,
		false,
	)
}

func (s *Service) Export() error {
	logger.Info(logModule,
		"exporting cleaned split file",
	)
	sfBytes, err := s.repository.GetLoadedSplitFile()
	if err != nil {
		logger.Errorf(logModule, "Failed to get loaded split file: %v", err)
		return err
	}

	sf, err := adapters.JSONSplitFileToDTO(string(sfBytes))
	if err != nil {
		logger.Errorf(logModule, "Failed to convert split file to domain object: %v", err)
		return err
	}

	cleanDTO, err := adapters.CleanSplitFile(sf)
	if err != nil {
		logger.Errorf(logModule, "Failed to clean user data from split file: %v", err)
		return err
	}

	cleanBytes, err := adapters.SplitFileToFrontEnd(cleanDTO)
	if err != nil {
		logger.Errorf(logModule, "Failed to convert clean split file to json: %v", err)
		return err
	}

	defaultFileName := buildSplitFileName(sf)
	logger.Infof(
		logModule,
		"export complete: %s",
		defaultFileName,
	)
	return s.repository.Export(cleanBytes, defaultFileName)
}

func (s *Service) Close() {
	s.splitFileLock.Lock()
	s.repository.ClearCachedFileName()
	s.splitFileLock.Unlock()
	logger.Infof(logModule, "repository cleared splitfile")
}

func (s *Service) SaveConfig(configService *config.Service) error {
	payload, err := adapters.ConfigToFrontEnd(configService)
	if err != nil {
		return err
	}

	logger.Debug(logModule, "repository saving config")
	s.configLock.Lock()
	err = s.repository.SaveConfig(payload)
	s.configLock.Unlock()

	if err != nil {
		logger.Errorf(logModule, "repo failed to save config: %s", err)
		return err
	}

	logger.Infof(logModule, "repository saved config")
	return nil
}

func (s *Service) LoadConfig(c *config.Service) error {
	logger.Debug(logModule, "repository loading config")
	s.configLock.RLock()
	b, err := s.repository.LoadConfig()
	if err != nil {
		s.configLock.RUnlock()
		logger.Errorf(logModule, "repo failed to load config: %s", err)
		return err
	}
	s.configLock.RUnlock()

	newConfig, err := adapters.FrontEndToConfig(b)
	if err != nil {
		return err
	}

	logger.Debug(
		logModule,
		"ensuring default key bindings",
	)
	newConfig.EnsureDefaultKeyBindings()

	s.configLock.Lock()
	c.SpeedRunAPIBase = newConfig.SpeedRunAPIBase
	c.KeyConfig = newConfig.KeyConfig
	c.GlobalHotkeysActive = newConfig.GlobalHotkeysActive
	c.SelectedSkin = newConfig.SelectedSkin
	c.SplitFileDir = newConfig.SplitFileDir
	c.SkinsDir = newConfig.SkinsDir
	s.configLock.Unlock()
	logger.Info(logModule, "repo loaded config")
	return nil
}
