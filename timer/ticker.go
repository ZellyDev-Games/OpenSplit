package timer

import "time"

type Ticker struct{ t *time.Ticker }

func NewTicker(d time.Duration) *Ticker { return &Ticker{time.NewTicker(d)} }

func (t Ticker) Ch() <-chan time.Time { return t.t.C }
func (t Ticker) Stop()                { t.t.Stop() }
