package persister

import (
	"OpenSplit/session"
	"context"
)

type Persister interface {
	Save(*session.SplitFile) error
	Load() (*session.SplitFile, error)
	Startup(context.Context)
}

type Service struct {
	ctx       context.Context
	persister Persister
	splitFile *session.SplitFile
}

func NewService(p Persister) *Service {
	return &Service{persister: p}
}

func (s *Service) Startup(ctx context.Context) {
	s.ctx = ctx
	s.persister.Startup(ctx)
}

func (s *Service) SetSplitFile(file *session.SplitFile) {
	s.splitFile = file
}

func (s *Service) Save() error {
	return s.persister.Save(s.splitFile)
}

func (s *Service) Load() (*session.SplitFile, error) {
	return s.Load()
}
