package repo

import (
	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/repo/adapters"
	"github.com/zellydev-games/opensplit/session"
)

// Repository defines a contract for a repo provider to operate against
type Repository interface {
	Load() ([]byte, error)
	Save([]byte) error
	SaveAs([]byte) error
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Load() (*session.SplitFile, error) {
	splitFile, err := s.repository.Load()
	if err != nil {
		return nil, err
	}
	dto, _ := adapters.JsonToSplitFile(string(splitFile))
	return adapters.SplitFileToDomain(dto)
}

func (s *Service) Save(splitFile *dto.SplitFile, X int, Y int, Width int, Height int) error {
	splitFile.WindowX = X
	splitFile.WindowY = Y
	splitFile.WindowWidth = Width
	splitFile.WindowHeight = Height
	payload, err := adapters.SplitFileToJson(splitFile)
	if err != nil {
		return err
	}
	return s.repository.Save(payload)
}
