package hotkeys

import (
	"testing"

	"github.com/zellydev-games/opensplit/dispatcher"
)

type MockHotkeyProvider struct {
	StartHookCalled int
	UnhookCalled    int
}

type MockDispatcher struct {
	DispatchCalled int
}

func (m *MockDispatcher) Dispatch(command dispatcher.Command, payload *string) (dispatcher.DispatchReply, error) {
	m.DispatchCalled++
	return dispatcher.DispatchReply{}, nil
}

func (m *MockHotkeyProvider) StartHook() error {
	m.StartHookCalled++
	return nil
}

func (m *MockHotkeyProvider) Unhook() error {
	m.UnhookCalled++
	return nil
}

func TestDispatcher(t *testing.T) {
	ch := make(chan KeyInfo)
	stateMachine := &MockDispatcher{}
	hotkeyProvider := &MockHotkeyProvider{}
	hotkeyService := NewService(ch, stateMachine, hotkeyProvider)

	hotkeyService.StartDispatcher()
	ch <- KeyInfo{KeyCode: 32}
	hotkeyService.StopDispatcher()
	if stateMachine.DispatchCalled != 1 {
		t.Errorf("command.SPLIT was not dispatched")
	}

	if hotkeyProvider.StartHookCalled != 1 {
		t.Errorf("hotkeyProvider.StartHook was not called")
	}

	if hotkeyProvider.UnhookCalled != 1 {
		t.Errorf("hotkeyProvider.Unhook was not called")
	}
}
