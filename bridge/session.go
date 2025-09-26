package bridge

import (
	"github.com/zellydev-games/opensplit/repo/adapters"
	"github.com/zellydev-games/opensplit/session"
)

type Session struct {
	runtimeProvider       RuntimeProvider
	sessionUpdatedChannel chan *session.Service
}

func NewSession(sessionUpdatedChannel chan *session.Service, runtimeProvider RuntimeProvider) *Session {
	return &Session{
		runtimeProvider:       runtimeProvider,
		sessionUpdatedChannel: sessionUpdatedChannel,
	}
}

func (s *Session) StartUIPump() {
	go func() {
		for updatedSession := range s.sessionUpdatedChannel {
			s.runtimeProvider.EventsEmit("session:update", adapters.DomainToSession(updatedSession))
		}
	}()
}
