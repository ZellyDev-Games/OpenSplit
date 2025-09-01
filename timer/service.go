package timer

import (
	"context"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Service struct {
	ctx         context.Context
	currentTime time.Duration
	startTime   time.Time
	running     bool
	mu          sync.Mutex
}

func NewService() *Service {
	return &Service{
		ctx:         nil,
		currentTime: 0,
		running:     false,
	}
}

func (s *Service) Startup(ctx context.Context) {
	s.ctx = ctx
	s.Run()
}

func (s *Service) IsRunning() bool {
	return s.running
}

func (s *Service) Run() {
	ticker := time.NewTicker(17 * time.Millisecond) // update ~60fps
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				s.mu.Lock()
				if s.running {
					// elapsed since last Start + whatever we had before
					s.currentTime = time.Since(s.startTime)
					runtime.EventsEmit(s.ctx, "timer:update", s.currentTime.Milliseconds())
				}
				s.mu.Unlock()
			}
		}
	}()
}

func (s *Service) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		// mark base time relative to now
		s.startTime = time.Now().Add(-s.currentTime)
		s.running = true
	}
}

func (s *Service) Pause() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		s.currentTime = time.Since(s.startTime)
		s.running = false
	}
}

func (s *Service) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
	s.currentTime = 0
}
