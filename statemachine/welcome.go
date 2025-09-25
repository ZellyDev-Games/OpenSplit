package statemachine

import (
	"fmt"

	"github.com/zellydev-games/opensplit/logger"
)

// Welcome greets the user by indicating the frontend should display the Welcome screen
type Welcome struct{}

// NewWelcomeState requires a *session.Service as its one and only payload parameter
func NewWelcomeState() (*Welcome, error) {
	return &Welcome{}, nil
}

// String returns the human friendly name for a State
func (w *Welcome) String() string {
	return "Welcome"
}

// OnEnter sets the context from the Wails app and signals the frontend to show the Welcome component
func (w *Welcome) OnEnter() error {
	machine.runtimeProvider.EventsEmit("state:enter", WELCOME)
	return nil
}
func (w *Welcome) OnExit() error { return nil }
func (w *Welcome) Receive(command Command, payload *string) (DispatchReply, error) {
	switch command {
	case LOAD:
		logger.Debug("Welcome received command LOAD")
		sf, err := machine.repoService.Load()
		if err != nil {
			return DispatchReply{1, "failed to load dto: " + err.Error()}, err
		}
		machine.sessionService.SetLoadedSplitFile(sf)
		machine.changeState(RUNNING)
		return DispatchReply{}, nil
	case NEW:
		logger.Debug("Welcome received command NEW")
		machine.changeState(NEWFILE)
		return DispatchReply{}, nil
	default:
		return DispatchReply{}, fmt.Errorf("invalid command %d for state Welcome", command)
	}
}
