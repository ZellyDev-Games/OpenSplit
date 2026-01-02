package bridge

import (
	"reflect"
	"testing"
	"time"

	"github.com/zellydev-games/opensplit/repo/adapters"
	"github.com/zellydev-games/opensplit/session"
)

type sessionEmitCall struct {
	event string
	args  []any
}

func TestNewSession(t *testing.T) {
	c := make(chan *session.Service)
	rp := newMockRuntimeProvider()
	s := NewSession(c, rp)

	if s.sessionUpdatedChannel != c || s.runtimeProvider != rp {
		t.Fatal("NewSession() did not return a new session with given parameters")
	}
}

func TestStartUIPump(t *testing.T) {
	rp := newMockRuntimeProvider()

	updates := make(chan *session.Service, 1)
	s := &Session{
		runtimeProvider:       rp,
		sessionUpdatedChannel: updates,
	}

	s.StartUIPump()

	domainSession := &session.Service{}
	expectedDTO := adapters.DomainToDTO(domainSession)

	updates <- domainSession

	select {
	case call := <-rp.signal:
		if call.event != "session:update" {
			t.Fatalf("expected event %q, got %q", "session:update", call.event)
		}

		if len(call.args) != 1 {
			t.Fatalf("expected 1 arg, got %d", len(call.args))
		}

		if !reflect.DeepEqual(call.args[0], expectedDTO) {
			t.Fatalf(
				"expected DTO %#v, got %#v",
				expectedDTO,
				call.args[0],
			)
		}

	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for EventsEmit")
	}
}

func TestStartUIPumpStops(t *testing.T) {
	rp := newMockRuntimeProvider()

	updates := make(chan *session.Service, 1)
	s := &Session{
		runtimeProvider:       rp,
		sessionUpdatedChannel: updates,
	}

	s.StartUIPump()

	updates <- &session.Service{}

	select {
	case <-rp.signal:
		// ok
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected initial EventsEmit")
	}

	close(updates)

	select {
	case call := <-rp.signal:
		t.Fatalf("unexpected EventsEmit after channel close: %#v", call)
	case <-time.After(100 * time.Millisecond):
		// ok
	}

	if len(rp.Calls()) != 1 {
		t.Fatalf("expected exactly one EventsEmit call, got %d", len(rp.Calls()))
	}
}

func TestEmitUIEvent(t *testing.T) {
	rp := newMockRuntimeProvider()
	model := AppViewModel{}
	EmitUIEvent(rp, model)

	select {
	case call := <-rp.signal:
		if call.event != uiModelEventName {
			t.Fatalf("expected event %q, got %q", uiModelEventName, call.event)
		}

		if len(call.args) != 1 {
			t.Fatalf("expected 1 arg, got %d", len(call.args))
		}

		got, ok := call.args[0].(AppViewModel)
		if !ok {
			t.Fatalf("expected AppViewModel arg, got %T", call.args[0])
		}

		if got != model {
			t.Fatalf("expected model %#v, got %#v", model, got)
		}

	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for EventsEmit")
	}
}
