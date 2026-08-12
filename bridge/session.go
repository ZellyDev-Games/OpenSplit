package bridge

import (
	"github.com/zellydev-games/opensplit/config"
	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/repo/adapters"
	"github.com/zellydev-games/opensplit/session"
)

const uiModelEventName = "ui:model"

// Session forwards session updates to the frontend.
type Session struct {
	runtimeProvider       RuntimeProvider
	sessionUpdatedChannel chan *session.Service
}

func NewSession(
	sessionUpdatedChannel chan *session.Service,
	runtimeProvider RuntimeProvider,
) *Session {
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
				logger.Debug(logModule, "session UI pump stopped")
				return
			}

			s.runtimeProvider.EventsEmit(
				"session:update",
				adapters.DomainToDTO(updatedSession),
			)
		}
	}()

	logger.Debug(logModule, "session UI pump started")
}

type WindowConfig struct {
	Width  int `json:"width"`
	Height int `json:"height"`

	Resizable bool `json:"resizable"`

	Positioned bool `json:"positioned,omitempty"`

	X int `json:"x,omitempty"`
	Y int `json:"y,omitempty"`
}

// View identifies the current application screen.
type View string

const (
	AppViewWelcome       View = "welcome"
	AppViewNewSplitFile  View = "new-split-file"
	AppViewEditSplitFile View = "edit-split-file"

	// Skin workflow
	AppViewNewSkin    View = "new-skin"
	AppViewEditSkin   View = "edit-skin"
	AppViewSkinEditor View = "skin-editor"

	AppViewRunning  View = "running"
	AppViewSettings View = "settings"
)

// AppViewModel describes the UI state presented to the frontend.
type AppViewModel struct {
	View View `json:"view"`

	Window WindowConfig `json:"window"`

	// Split editor
	SplitFile *dto.SplitFile `json:"splitFile,omitempty"`

	// Running timer
	Session *dto.Session `json:"session,omitempty"`

	// Settings
	Config *config.Service `json:"config,omitempty"`

	// Skin editor
	Skin any `json:"skin,omitempty"`

	// Available installed skins.
	//
	// Used by:
	//   • Edit Skin dropdown
	//   • New Skin base-skin dropdown
	AvailableSkins []string `json:"availableSkins,omitempty"`

	// Selected skin name.
	//
	// Used when editing an existing skin.
	SelectedSkin string `json:"selectedSkin,omitempty"`
}

// EmitUIEvent informs the frontend of a state change.
func EmitUIEvent(runtimeProvider RuntimeProvider, model AppViewModel) {
	logger.Debugf(logModule, "setting UI model: %s", model.View)

	model.Window = WindowForView(model.View)

	if model.View == AppViewRunning &&
		model.Session != nil &&
		model.Session.LoadedSplitFile != nil {

		splitFile := model.Session.LoadedSplitFile

		var window dto.SplitterWindow

		switch splitFile.Layout {
		case "horizontal":
			window = splitFile.Windows.Horizontal
		default:
			window = splitFile.Windows.Vertical
		}

		model.Window.Width = window.Width
		model.Window.Height = window.Height
		model.Window.X = window.X
		model.Window.Y = window.Y
		model.Window.Positioned = true
		model.Window.Resizable = true
	}

	runtimeProvider.EventsEmit(uiModelEventName, model)
}
