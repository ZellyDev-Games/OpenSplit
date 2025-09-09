package hotkeys

import (
	"testing"
)

type MockHotkeyProvider struct {
	StartHookCalled int
	UnhookCalled    int
}

func (m *MockHotkeyProvider) StartHook() error {
	m.StartHookCalled++
	return nil
}

func (m *MockHotkeyProvider) Unhook() error {
	m.UnhookCalled++
	return nil
}

type MockSession struct {
	SplitCalled int
}

func (m *MockSession) Split() {
	m.SplitCalled++
}

func TestDispatcher(t *testing.T) {
	ch := make(chan KeyInfo)
	session := &MockSession{}
	hotkeyProvider := &MockHotkeyProvider{}
	service := NewService(ch, session, hotkeyProvider)

	service.StartDispatcher()
	ch <- KeyInfo{KeyCode: 32}
	service.StopDispatcher()
	if session.SplitCalled != 1 {
		t.Errorf("session.Split was not called")
	}

	if hotkeyProvider.StartHookCalled != 1 {
		t.Errorf("hotkeyProvider.StartHook was not called")
	}

	if hotkeyProvider.UnhookCalled != 1 {
		t.Errorf("hotkeyProvider.Unhook was not called")
	}
}
