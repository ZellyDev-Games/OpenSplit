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
	mockT.ch = make(chan time.Time, 1)
	s, timeUpdatedChannel := NewStopwatch(mockT)
	ctx, cancel := context.WithCancel(context.Background())
	s.Startup(ctx)

	go func() {
		select {
		case got := <-timeUpdatedChannel:
			if got != 42*time.Millisecond {
				t.Errorf("time updated: got %v, want %v", got, 42*time.Millisecond)
			}
		case <-time.After(time.Millisecond * 100):
			t.Errorf("channel was not sent to")
		}
	}()

	s.Run()
	mockT.ch <- time.Unix(0, 42e6)
	cancel()
}

func TestFormatTimeToString(t *testing.T) {
	d := time.Hour*1 + time.Minute*2 + time.Second*3 + time.Millisecond*400
	timeString := FormatTimeToString(d)
	if timeString != "01:02:03.40" {
		t.Errorf("FormatTimeToString() got %s, want %s", timeString, "01:02:03.40")
	}

	d = time.Minute*2 + time.Second*3 + time.Millisecond*400
	timeString = FormatTimeToString(d)
	if timeString != "00:02:03.40" {
		t.Errorf("FormatTimeToString() got %s, want %s", timeString, "00:02:03.40")
	}
}

func TestParseStringToTime(t *testing.T) {
	timeString := "1:02:03.40"
	d, _ := ParseStringToTime(timeString)
	w := time.Hour*1 + time.Minute*2 + time.Second*3 + time.Millisecond*400
	if d != w {
		t.Errorf("ParseStringToTime() got %d, want %d", d, w)
	}
}
