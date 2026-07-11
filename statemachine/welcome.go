package statemachine

import (
	"fmt"

	"github.com/zellydev-games/opensplit/bridge"
	"github.com/zellydev-games/opensplit/command"
	"github.com/zellydev-games/opensplit/dispatcher"
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/repo/adapters"
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
	bridge.EmitUIEvent(machine.runtimeProvider, bridge.AppViewModel{
		View: bridge.AppViewWelcome,
	})
	return nil
}
func (w *Welcome) OnExit() error { return nil }
func (w *Welcome) Receive(c command.Command, _ *string) (dispatcher.DispatchReply, error) {
	switch c {
	case command.LOAD:
		logger.Debug(logModule, "Welcome received c LOAD")
		sf, err := machine.repoService.LoadSplitFile()
		if err != nil {
			return dispatcher.DispatchReply{Code: 1, Message: "failed to load dto: " + err.Error()}, err
		}
		if sf.SelectedSkin != "" {
			if err := machine.skinProvider.SetSkin(sf.SelectedSkin, false); err != nil {
				logger.Errorf(logModule, "failed to set skin: %v", err)
			}
		} else {
			if err := machine.skinProvider.SetSkin(machine.configService.SelectedSkin, false); err != nil {
				logger.Errorf(logModule, "failed to set skin: %v", err)
			}
		}
		domainSF, err := adapters.DTOSplitFileToDomain(sf)
		if err != nil {
			return dispatcher.DispatchReply{Code: 2, Message: "failed to convert dto: " + err.Error()}, err
		}
		machine.sessionService.SetLoadedSplitFile(domainSF)
		go machine.updateWorldRecord()
		machine.changeState(RUNNING)
		return dispatcher.DispatchReply{}, nil
	case command.NEW:
		logger.Debug(logModule, "Welcome received c NEW")
		machine.changeState(NEWFILE)
		return dispatcher.DispatchReply{}, nil
	case command.EDIT:
		logger.Debug(logModule, "Welcome received c EDIT")
		machine.changeState(CONFIG)
		return dispatcher.DispatchReply{}, nil
	default:
		return dispatcher.DispatchReply{}, fmt.Errorf("invalid c %d for state Welcome", c)
	}
}
