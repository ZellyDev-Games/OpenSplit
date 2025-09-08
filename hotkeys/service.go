package hotkeys

import (
	"OpenSplit/logger"
	"OpenSplit/session"
	"fmt"
)

type HotkeyProvider interface {
	StartHook() error
	Unhook() error
}

type KeyInfo struct {
	KeyCode    int
	LocaleName string
}

type Service struct {
	hotkeyChannel  chan KeyInfo
	hotkeyProvider HotkeyProvider
	sessionService *session.Service
}

func NewService(keyInfoChannel chan KeyInfo, sessionService *session.Service, provider HotkeyProvider) *Service {
	return &Service{
		hotkeyChannel:  keyInfoChannel,
		sessionService: sessionService,
		hotkeyProvider: provider,
	}
}

func (s *Service) StartDispatcher() {
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
}

func (s *Service) dispatch() {
	for {
		select {
		case keyInfo := <-s.hotkeyChannel:
			switch keyInfo.KeyCode {
			case 32:
				s.sessionService.Split()
			}
		}
	}
}
