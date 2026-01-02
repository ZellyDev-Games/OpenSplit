package bridge

import "sync"

type emitCall struct {
	event string
	args  []any
}

type mockRuntimeProvider struct {
	mu     sync.Mutex
	calls  []emitCall
	signal chan emitCall
}

func newMockRuntimeProvider() *mockRuntimeProvider {
	return &mockRuntimeProvider{
		signal: make(chan emitCall, 16),
	}
}

func (m *mockRuntimeProvider) EventsEmit(event string, args ...any) {
	call := emitCall{event: event, args: args}

	m.mu.Lock()
	m.calls = append(m.calls, call)
	m.mu.Unlock()

	// Notify test without blocking (buffered).
	select {
	case m.signal <- call:
	default:
	}
}

func (m *mockRuntimeProvider) Calls() []emitCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]emitCall, len(m.calls))
	copy(out, m.calls)
	return out
}
