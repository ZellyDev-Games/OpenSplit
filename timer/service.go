package timer

import (
	"context"
	"strconv"
	"sync"
	"time"
)

type TickerInterface interface {
	Ch() <-chan time.Time
	Stop()
}

type Service struct {
	ctx                context.Context
	currentTime        time.Duration
	startTime          time.Time
	running            bool
	mu                 sync.Mutex
	timeUpdatedChannel chan time.Duration
	ticker             TickerInterface
}

func NewService(t TickerInterface) (*Service, chan time.Duration) {
	timeUpdatedChannel := make(chan time.Duration)
	return &Service{
		ctx:                nil,
		currentTime:        0,
		running:            false,
		timeUpdatedChannel: timeUpdatedChannel,
		ticker:             t,
	}, timeUpdatedChannel
}

func (s *Service) Startup(ctx context.Context) {
	s.ctx = ctx
	s.Run()
}

func (s *Service) IsRunning() bool {
	return s.running
}

func (s *Service) Run() {
	go func() {
		defer s.ticker.Stop()
		for {
			select {
			case <-s.ctx.Done():
				return
			case t := <-s.ticker.Ch():
				s.tickOnce(t)
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

func (s *Service) GetCurrentTimeFormatted() string {
	return strconv.FormatInt(s.currentTime.Milliseconds(), 10)
}

func (s *Service) GetCurrentTime() time.Duration {
	return s.currentTime
}

func (s *Service) tickOnce(now time.Time) {
	if s.running {
		s.mu.Lock()
		s.currentTime = now.Sub(s.startTime)
		s.mu.Unlock()
		s.timeUpdatedChannel <- s.currentTime
	}
}
