package hotkeys

import (
	"OpenSplit/logger"
	"fmt"
)

type HotkeyProvider interface {
	StartHook() error
	Unhook() error
}

type Splitter interface {
	Split()
}

type KeyInfo struct {
	KeyCode    int
	LocaleName string
}

type Service struct {
	hotkeyChannel  chan KeyInfo
	hotkeyProvider HotkeyProvider
	sessionService Splitter
	internalStop   chan struct{}
}

func NewService(keyInfoChannel chan KeyInfo, sessionService Splitter, provider HotkeyProvider) *Service {
	return &Service{
		hotkeyChannel:  keyInfoChannel,
		sessionService: sessionService,
		hotkeyProvider: provider,
	}
}

func (s *Service) StartDispatcher() {
	s.internalStop = make(chan struct{})
	err := s.hotkeyProvider.StartHook()
	if err != nil {
		logger.Error(fmt.Sprintf("failed to add hotkey provider hook: %s", err))
	}
	go s.dispatch()
}

func (s *Service) StopDispatcher() {
	err := s.hotkeyProvider.Unhook()
	if err != nil {
		logger.Error(fmt.Sprintf("failed to unhook hotkey provider: %s", err))
	}
	close(s.internalStop)
}

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
