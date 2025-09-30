package statemachine

import (
	"fmt"

	"github.com/zellydev-games/opensplit/dispatcher"
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/repo/adapters"
)

// Running represents the state where a dto has been loaded, the UI should be showing the SplitList and the timer.
type Running struct{}

func NewRunningState() (*Running, error) {
	return &Running{}, nil
}

func (r Running) OnEnter() error {
	sessionDto := adapters.DomainToSession(machine.sessionService)

	if machine.hotkeyProvider != nil {
		machine.hotkeyProvider.StartDispatcher()
	}

	machine.runtimeProvider.EventsEmit("state:enter", RUNNING, sessionDto)
	return nil
}

func (r Running) OnExit() error {
	if machine.hotkeyProvider != nil {
		machine.hotkeyProvider.StopDispatcher()
	}
	return nil
}

func (r Running) Receive(command dispatcher.Command, params *string) (dispatcher.DispatchReply, error) {
	switch command {
	case dispatcher.CLOSE:
		logger.Debug(fmt.Sprintf("Running received CLOSE command: %v", params))
		machine.sessionService.CloseRun()
		machine.repoService.Close()
		machine.changeState(WELCOME, nil)
	case dispatcher.EDIT:
		logger.Debug(fmt.Sprintf("Running received EDIT command: %v", params))
		machine.changeState(EDITING, nil)
	case dispatcher.SAVE:
		logger.Debug(fmt.Sprintf("Running received SAVE command: %v", params))
		sf := machine.sessionService.SplitFile()
		w, h := machine.runtimeProvider.WindowGetSize()
		x, y := machine.runtimeProvider.WindowGetPosition()
		dto := adapters.DomainToSplitFile(sf)
		err := machine.repoService.SaveSplitFile(dto, x, y, w, h)
		if err != nil {
			msg := fmt.Sprintf("failed to save split file to session: %s", err)
			logger.Error(msg)
			return dispatcher.DispatchReply{Code: 2, Message: msg}, err
		}
	case dispatcher.SPLIT:
		logger.Debug(fmt.Sprintf("Running received SPLIT command: %v", params))
		machine.sessionService.Split()
		return dispatcher.DispatchReply{}, nil
	default:
		panic("unhandled default case in Running")
	}

	return dispatcher.DispatchReply{}, nil
}

func (r Running) String() string {
	return "Running"
}
