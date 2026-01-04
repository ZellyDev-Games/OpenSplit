package repo

import (
	"errors"

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
	SaveSplitFile([]byte, string) error
	SaveAs([]byte, string) error
	ClearCachedFileName()
	SaveConfig([]byte) error
	LoadConfig() ([]byte, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) LoadSplitFile() (session.SplitFile, error) {
	splitFile, err := s.repository.LoadSplitFile()
	if err != nil {
		return session.SplitFile{}, err
	}
	splitFileSTO, _ := adapters.JSONSplitFileToDTO(string(splitFile))
	return adapters.DTOSplitFileToDomain(splitFileSTO)
}

func (s *Service) SaveSplitFile(splitFile dto.SplitFile, X int, Y int, Width int, Height int) error {
	splitFile.WindowX = X
	splitFile.WindowY = Y
	splitFile.WindowWidth = Width
	splitFile.WindowHeight = Height
	payload, err := adapters.SplitFileToFrontEnd(splitFile)
	if err != nil {
		return err
	}
	identifier := splitFile.GameName
	if splitFile.GameCategory != "" {
		identifier += "-" + splitFile.GameCategory
	}
	identifier += ".osf"
	return s.repository.SaveSplitFile(payload, identifier)
}

func (s *Service) Close() {
	s.repository.ClearCachedFileName()
}

func (s *Service) SaveConfig(configService *config.Service) error {
	payload, err := adapters.ConfigToFrontEnd(configService)
	if err != nil {
		return err
	}
	return s.repository.SaveConfig(payload)
}

func (s *Service) LoadConfig(c *config.Service) error {
	b, err := s.repository.LoadConfig()
	if err != nil {
		return err
	}

	newConfig, err := adapters.FrontEndToConfig(b)
	if err != nil {
		return err
	}

	c.SpeedRunAPIBase = newConfig.SpeedRunAPIBase
	c.KeyConfig = newConfig.KeyConfig
	return nil
}
