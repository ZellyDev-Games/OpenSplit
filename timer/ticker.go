package timer

import "time"

// Ticker wraps time.Ticker to allow DI for testing
type Ticker struct{ t *time.Ticker }

// NewTicker constructor
func NewTicker(d time.Duration) *Ticker { return &Ticker{time.NewTicker(d)} }

// Ch provides read access to the channel that the ticker uses to signal ticks.
func (t Ticker) Ch() <-chan time.Time { return t.t.C }

// Stop terminates ticks.
func (t Ticker) Stop() { t.t.Stop() }
