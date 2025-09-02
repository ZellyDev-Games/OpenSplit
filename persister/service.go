package persister

import (
	"OpenSplit/session"
	"context"
)

type Persister interface {
	Save(payload session.SplitFilePayload) error
	Load() (session.SplitFilePayload, error)
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

func (s *Service) Save(payload session.SplitFilePayload) error {
	return s.persister.Save(payload)
}

func (s *Service) Load() (session.SplitFilePayload, error) {
	return s.Load()
}
