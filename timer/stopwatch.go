package timer

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"
)

// TickerInterface wraps time.Ticker to allow DI for testing
type TickerInterface interface {
	Ch() <-chan time.Time
	Stop()
}

// Stopwatch tracks the passing of time with a monotonic clock
//
// Stopwatch increments the currentTime anytime the attached TickerInterface ticks and running is true.
// Stopwatch uses Go's monotonic clock to generate diffs between the startTime and ticker current time to increment currentTime
type Stopwatch struct {
	ctx                context.Context
	currentTime        time.Duration
	startTime          time.Time
	running            bool
	mu                 sync.Mutex
	timeUpdatedChannel chan time.Duration
	ticker             TickerInterface
}

// NewStopwatch returns a Stopwatch and a channel that sends the currentTime at approximately 60 FPS.
func NewStopwatch(t TickerInterface) (*Stopwatch, chan time.Duration) {
	timeUpdatedChannel := make(chan time.Duration)
	return &Stopwatch{
		ctx:                nil,
		currentTime:        0,
		running:            false,
		timeUpdatedChannel: timeUpdatedChannel,
		ticker:             t,
	}, timeUpdatedChannel
}

// Startup receives a context.Context from Wails.run
func (s *Stopwatch) Startup(ctx context.Context) {
	s.ctx = ctx
	s.Run()
}

// IsRunning returns the running state of the Timer
func (s *Stopwatch) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *Stopwatch) Run() {
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
func (s *Stopwatch) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		// mark base time relative to now
		s.startTime = time.Now().Add(-s.currentTime)
		s.running = true
	}
}

// Pause freezes the current time, and sets the running state to false so ticker updates stop accumulating.
func (s *Stopwatch) Pause() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		s.currentTime = time.Since(s.startTime)
		s.running = false
	}
}

// Reset sets the running state to false, stopping the ticker updates from accumulating then sets the current time to 0.
func (s *Stopwatch) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
	s.currentTime = 0
	select {
	case s.timeUpdatedChannel <- 0:
	default:
	}
}

// GetCurrentTimeFormatted returns a frontend friendly string representing the current accumulated time.
func (s *Stopwatch) GetCurrentTimeFormatted() string {
	return strconv.FormatInt(s.currentTime.Milliseconds(), 10)
}

// GetCurrentTime allows public read access to the currentTime
func (s *Stopwatch) GetCurrentTime() time.Duration {
	return s.currentTime
}

// FormatTimeToString takes a time.Duration and returns a string designed to be worked with by the frontend.
//
// Inverse of ParseStringToTime
func FormatTimeToString(d time.Duration) string {
	sign := ""
	if d < 0 {
		sign = "-"
		d = -d
	}
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	cs := (d - s*time.Second) / (10 * time.Millisecond) // centiseconds

	return fmt.Sprintf("%s%02d:%02d:%02d.%02d", sign, h, m, s, cs)
}

// ParseStringToTime unserializes a string, usually from the frontend, into a time.Duration.
//
// Inverse of FormatTimeToString.
func ParseStringToTime(s string) (time.Duration, error) {
	var neg bool
	if len(s) > 0 && s[0] == '-' {
		neg = true
		s = s[1:]
	}

	var h, m, sec, cs int
	_, err := fmt.Sscanf(s, "%02d:%02d:%02d.%02d", &h, &m, &sec, &cs)
	if err != nil {
		return 0, err
	}

	d := (time.Duration(h) * time.Hour) +
		(time.Duration(m) * time.Minute) +
		(time.Duration(sec) * time.Second) +
		(time.Duration(cs) * 10 * time.Millisecond)

	if neg {
		d = -d
	}
	return d, nil
}

func PayloadRawTimeToDuration(ms int64) time.Duration {
	return time.Duration(ms) * time.Millisecond
}

func (s *Stopwatch) tickOnce(now time.Time) {
	if s.running {
		s.mu.Lock()
		s.currentTime = now.Sub(s.startTime)
		s.mu.Unlock()
		select {
		case s.timeUpdatedChannel <- s.currentTime:
		default:
			fmt.Println("Failed to sent to timeUpdatedChannel")
		}
	}
}

// Ticker wraps time.Ticker to allow DI for testing
type Ticker struct{ t *time.Ticker }

// NewTicker constructor
func NewTicker(d time.Duration) *Ticker { return &Ticker{time.NewTicker(d)} }

// Ch provides read access to the channel that the ticker uses to signal ticks.
func (t Ticker) Ch() <-chan time.Time { return t.t.C }

// Stop terminates ticks.
func (t Ticker) Stop() { t.t.Stop() }
