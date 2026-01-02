package bridge

import (
	"testing"
	"time"

	"github.com/zellydev-games/opensplit/config"
)

func TestStartUIPumpEmitsConfigUpdate(t *testing.T) {
	rp := newMockRuntimeProvider()

	updates := make(chan *config.Service, 1)
	c := NewConfig(updates, rp)

	c.StartUIPump()

	want := &config.Service{}
	updates <- want

	select {
	case call := <-rp.signal:
		if call.event != "config:update" {
			t.Fatalf("expected event %q, got %q", "config:update", call.event)
		}
		if len(call.args) != 1 {
			t.Fatalf("expected 1 arg, got %d (%#v)", len(call.args), call.args)
		}
		gotPtr, ok := call.args[0].(*config.Service)
		if !ok {
			t.Fatalf("expected arg[0] to be *config.Service, got %T", call.args[0])
		}
		if gotPtr != want {
			t.Fatalf("expected emitted *config.Service pointer %p, got %p", want, gotPtr)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for EventsEmit to be called")
	}
}

func TestStartUIPumpStopsWhenChannelClosed(t *testing.T) {
	rp := newMockRuntimeProvider()

	updates := make(chan *config.Service, 1)
	c := NewConfig(updates, rp)

	c.StartUIPump()

	// First update should emit.
	first := &config.Service{}
	updates <- first

	select {
	case <-rp.signal:
		// ok
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for first EventsEmit call")
	}

	// Close the channel; pump should exit.
	close(updates)

	// Assert no additional emits occur after close within a short window.
	select {
	case call := <-rp.signal:
		t.Fatalf("unexpected EventsEmit after channel close: %#v", call)
	case <-time.After(100 * time.Millisecond):
		// ok
	}

	// Optional: ensure exactly one call was made.
	calls := rp.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call total, got %d (%#v)", len(calls), calls)
	}
	if calls[0].event != "config:update" {
		t.Fatalf("unexpected event: %q", calls[0].event)
	}
	if len(calls[0].args) != 1 || calls[0].args[0] != first {
		t.Fatalf("unexpected args: %#v", calls[0].args)
	}
}
