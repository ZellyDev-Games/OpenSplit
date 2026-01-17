package statemachine

import (
	"errors"
	"fmt"

	"github.com/zellydev-games/opensplit/bridge"
	"github.com/zellydev-games/opensplit/dispatcher"
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/repo"
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

func (w *Welcome) ID() StateID {
	return WELCOME
}

// OnEnter sets the context from the Wails app and signals the frontend to show the Welcome component
func (w *Welcome) OnEnter() error {
	err := machine.repoService.LoadConfig(machine.configService)
	if err != nil {
		if errors.Is(err, repo.ErrConfigMissing) {
			machine.configService.CreateDefaultConfig()
			err = machine.repoService.SaveConfig(machine.configService)
			if err != nil {
				logger.Errorf(logModule, "failed to create default config: %s", err.Error())
				return err
			}
		} else {
			return err
		}
	}

	bridge.EmitUIEvent(machine.runtimeProvider, bridge.AppViewModel{
		View: bridge.AppViewWelcome,
	})
	return nil
}
func (w *Welcome) OnExit() error { return nil }
func (w *Welcome) Receive(command dispatcher.Command, _ *string) (dispatcher.DispatchReply, error) {
	switch command {
	case dispatcher.LOAD:
		logger.Debug(logModule, "Welcome received command LOAD")
		sf, err := machine.repoService.LoadSplitFile()
		if err != nil {
			return dispatcher.DispatchReply{Code: 1, Message: "failed to load dto: " + err.Error()}, err
		}
		machine.sessionService.SetLoadedSplitFile(sf)
		machine.changeState(RUNNING)
		return dispatcher.DispatchReply{}, nil
	case dispatcher.NEW:
		logger.Debug(logModule, "Welcome received command NEW")
		machine.changeState(NEWFILE)
		return dispatcher.DispatchReply{}, nil
	case dispatcher.EDIT:
		logger.Debug(logModule, "Welcome received command EDIT")
		machine.changeState(CONFIG)
		return dispatcher.DispatchReply{}, nil
	default:
		return dispatcher.DispatchReply{}, fmt.Errorf("invalid command %d for state Welcome", command)
	}
}
