package timer

import (
	"context"
	"testing"
	"time"
)

type mockTicker struct{ ch chan time.Time }

func (m *mockTicker) Ch() <-chan time.Time { return m.ch }
func (m *mockTicker) Stop()                {}

func TestRun(t *testing.T) {
	mockT := &mockTicker{}
	mockT.ch = make(chan time.Time)
	s, timeUpdatedChannel := NewService(mockT)
	ctx, cancel := context.WithCancel(context.Background())
	s.Startup(ctx)
	s.running = true
	base := time.Unix(0, 0)
	s.startTime = base
	mockT.ch <- time.Unix(0, 42e6)

	select {
	case got := <-timeUpdatedChannel:
		if got != 42*time.Millisecond {
			t.Errorf("time updated: got %v, want %v", got, 42*time.Millisecond)
		}
	case <-time.After(time.Millisecond * 100):
		t.Errorf("channel was not sent to")
	}

	cancel()
}
