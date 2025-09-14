package hotkeys

import (
	"OpenSplit/logger"
	"fmt"
)

// HotkeyProvider must be implemented by any OS specific hotkey system to be used by the hotkeys.Service
type HotkeyProvider interface {
	StartHook() error
	Unhook() error
}

// Splitter is generally implemented by the session.Service created by the application, but this interface is useful
// for testing purposes
type Splitter interface {
	Split()
}

// KeyInfo is the Go-friendly struct to capture key code and key name data from the OS
type KeyInfo struct {
	KeyCode    int
	LocaleName string
}

// Service holds a channel that retrieves KeyInfo, controls the provided HotkeyProvider with StartHook/Unhook,
// and calls exported functions on the provided Splitter (usually session.Service if not testing)
type Service struct {
	hotkeyChannel  chan KeyInfo
	hotkeyProvider HotkeyProvider
	sessionService Splitter
	internalStop   chan struct{}
}

// NewService creates a new hotkeys.Service that holds a chan KeyInfo, a reference to a Splitter (usually session.Service)
// a HotkeyProvider that sends raw keypresses from the OS to the keyInfoChannel
//
// The common pattern used in OpenSplit is to create a HotkeyProvider with a constructor func that also returns a
// chan KeyInfo it sends keypress information to, and use that as the first parameter to this constructor func.
func NewService(keyInfoChannel chan KeyInfo, sessionService Splitter, provider HotkeyProvider) *Service {
	return &Service{
		hotkeyChannel:  keyInfoChannel,
		sessionService: sessionService,
		hotkeyProvider: provider,
	}
}

// StartDispatcher creates an internal channel that shuts down the dispatch loop when closed, starts the HotkeyProvider
// OS Hook, and starts the dispatch loop that listens on hotkeyChannel for KeyInfo events
func (s *Service) StartDispatcher() {
	s.internalStop = make(chan struct{})
	err := s.hotkeyProvider.StartHook()
	if err != nil {
		logger.Error(fmt.Sprintf("failed to add hotkey provider hook: %s", err))
	}
	go s.dispatch()
}

// StopDispatcher unhooks the HotkeyProvider from the OS, and closes the internal stop channel to stop the dispatch loop
func (s *Service) StopDispatcher() {
	err := s.hotkeyProvider.Unhook()
	if err != nil {
		logger.Error(fmt.Sprintf("failed to unhook hotkey provider: %s", err))
	}
	close(s.internalStop)
}

// dispatch listens for the internalStop channel to close (generally via StopDispatcher), and listens for a hotkey
// on the hotkeyChannel from the HotkeyProvider. It then calls the appropriate functions on the session.Service
// when specific hotkeys are pressed
//
// TODO: Make hotkeys configurable
func (s *Service) dispatch() {
	for {
		select {
		case <-s.internalStop:
			return

		case keyInfo, ok := <-s.hotkeyChannel:
			if !ok {
				logger.Warn("hotkeyChannel closed")
				return
			}
			switch keyInfo.KeyCode {
			case 32: // Space
				s.sessionService.Split()
			}
		}
	}
}
