package bridge

import (
	"testing"
	"time"
)

func TestStartTimerUIPump(t *testing.T) {
	rp := newMockRuntimeProvider()

	updates := make(chan time.Duration, 1)
	timer := NewTimer(updates, rp)

	timer.StartUIPump()

	d := 1500 * time.Millisecond
	updates <- d

	select {
	case call := <-rp.signal:
		if call.event != "timer:update" {
			t.Fatalf("expected event %q, got %q", "timer:update", call.event)
		}
		if len(call.args) != 1 {
			t.Fatalf("expected 1 arg, got %d", len(call.args))
		}
		// StartUIPump emits currentTime.Milliseconds() which is int64
		got, ok := call.args[0].(int64)
		if !ok {
			t.Fatalf("expected arg[0] to be int64 (milliseconds), got %T", call.args[0])
		}
		want := d.Milliseconds()
		if got != want {
			t.Fatalf("expected milliseconds %d, got %d", want, got)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for EventsEmit")
	}
}

func TestTimerUIPumpStops(t *testing.T) {
	rp := newMockRuntimeProvider()

	updates := make(chan time.Duration, 2)
	timer := NewTimer(updates, rp)

	timer.StartUIPump()

	// Prove it emits at least once.
	updates <- 10 * time.Millisecond
	select {
	case <-rp.signal:
		// ok
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected initial EventsEmit")
	}

	// Stop the pump.
	timer.timerEventStopChannel <- struct{}{}

	// Drain any already-emitted notifications quickly (optional defensive drain).
drain:
	for {
		select {
		case <-rp.signal:
			// ignore
		default:
			break drain
		}
	}

	// After stop, further updates should not emit.
	updates <- 20 * time.Millisecond

	select {
	case call := <-rp.signal:
		t.Fatalf("unexpected EventsEmit after stop: %#v", call)
	case <-time.After(100 * time.Millisecond):
		// ok
	}
}
