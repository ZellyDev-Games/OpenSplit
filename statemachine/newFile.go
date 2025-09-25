package statemachine

import (
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/repo/adapters"
)

// NewFile indicates that the frontend should show the SplitEditor, and cannot send along a split file.
//
// CANCEL from NewFile should return to the Welcome state
type NewFile struct{}

func NewNewFileState() (*NewFile, error) {
	return &NewFile{}, nil
}

func (n *NewFile) String() string {
	return "NewFile"
}

func (n *NewFile) OnEnter() error {
	machine.runtimeProvider.EventsEmit("state:enter", NEWFILE)
	return nil
}
func (n *NewFile) OnExit() error { return nil }
func (n *NewFile) Receive(command Command, payload *string) (DispatchReply, error) {
	switch command {
	case CANCEL:
		machine.changeState(WELCOME)
	case SUBMIT:
		if payload == nil {
			return DispatchReply{
				Code:    1,
				Message: "nil payload received",
			}, nil
		}
		dto, err := adapters.FrontendToSplitFile(*payload)
		if err != nil {
			logger.Error(err.Error())
			return DispatchReply{2, err.Error()}, err
		}
		err = machine.repoService.Save(dto, 100, 100, 390, 550)
		if err != nil {
			return DispatchReply{4, "failed to save dto: " + err.Error()}, err
		}
		sf, err := adapters.SplitFileToDomain(dto)
		if err != nil {
			return DispatchReply{5, err.Error()}, err
		}
		machine.sessionService.SetLoadedSplitFile(sf)
		machine.changeState(RUNNING)
		return DispatchReply{}, nil
	default:
		panic("unhandled default case")
	}
	return DispatchReply{}, nil
}
