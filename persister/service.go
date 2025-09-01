package persister

import (
	"OpenSplit/splits"
	"context"
)

type Persister interface {
	Save(*splits.SplitFile) error
	Load() (*splits.SplitFile, error)
	Startup(context.Context)
}

type Service struct {
	ctx       context.Context
	persister Persister
	splitFile *splits.SplitFile
}

func NewService(p Persister) *Service {
	return &Service{persister: p}
}

func (s *Service) Startup(ctx context.Context) {
	s.ctx = ctx
	s.persister.Startup(ctx)
}

func (s *Service) SetSplitFile(file *splits.SplitFile) {
	s.splitFile = file
}

func (s *Service) Save() error {
	return s.persister.Save(s.splitFile)
}

func (s *Service) Load() (*splits.SplitFile, error) {
	return s.Load()
}
