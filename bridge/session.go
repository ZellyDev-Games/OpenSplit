package bridge

import (
	"github.com/zellydev-games/opensplit/config"
	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/repo/adapters"
	"github.com/zellydev-games/opensplit/session"
)

const uiModelEventName = "ui:model"

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
		for {
			updatedSession, ok := <-s.sessionUpdatedChannel
			if !ok {
				return
			}
			s.runtimeProvider.EventsEmit("session:update", adapters.DomainToDTO(updatedSession))
		}
	}()
}

type View string

const (
	AppViewWelcome       View = "welcome"
	AppViewNewSplitFile  View = "new-split-file"
	AppViewEditSplitFile View = "edit-split-file"
	AppViewRunning       View = "running"
	AppViewSettings      View = "settings"
)

type AppViewModel struct {
	View View `json:"view"`

	// Only set for editor screens
	SpeedrunAPIBaseURL string         `json:"speedrunApiBaseUrl,omitempty"`
	SplitFile          *dto.SplitFile `json:"splitFile,omitempty"`

	// Only set for running
	Session *dto.Session `json:"session,omitempty"`

	// Only set for settings
	Config *config.Service `json:"config,omitempty"`
}

// EmitUIEvent informs the frontend of a state change
func EmitUIEvent(runtimeProvider RuntimeProvider, model AppViewModel) {
	runtimeProvider.EventsEmit(uiModelEventName, model)
}
