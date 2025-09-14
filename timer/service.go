package timer

import (
	"context"
	"strconv"
	"sync"
	"time"
)

// TickerInterface wraps time.Ticker to allow DI for testing
type TickerInterface interface {
	Ch() <-chan time.Time
	Stop()
}

// Service is a stopwatch
//
// Service increments the currentTime anytime the attached TickerInterface ticks and running is true.
// Service uses Go's monotonic clock to generate diffs between the startTime and ticker current time to increment currentTime
type Service struct {
	ctx                context.Context
	currentTime        time.Duration
	startTime          time.Time
	running            bool
	mu                 sync.Mutex
	timeUpdatedChannel chan time.Duration
	ticker             TickerInterface
}

// NewService returns a Service and a channel that sends the currentTime at approximately 60 FPS.
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

// Startup receives a context.Context from Wails.Run
func (s *Service) Startup(ctx context.Context) {
	s.ctx = ctx
	s.Run()
}

// IsRunning returns the running state of the Timer
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

// Start marks the current time for the monotonic clock and sets the running state to true
func (s *Service) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		// mark base time relative to now
		s.startTime = time.Now().Add(-s.currentTime)
		s.running = true
	}
}

// Pause freezes the current time, and sets the running state to false so ticker updates stop accumulating.
func (s *Service) Pause() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		s.currentTime = time.Since(s.startTime)
		s.running = false
	}
}

// Reset sets the running state to false, stopping the ticker updates from accumulating then sets the current time to 0.
func (s *Service) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
	s.currentTime = 0
}

// GetCurrentTimeFormatted returns a frontend friendly string representing the current accumulated time.
func (s *Service) GetCurrentTimeFormatted() string {
	return strconv.FormatInt(s.currentTime.Milliseconds(), 10)
}

// GetCurrentTime allows public read access to the currentTime
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
