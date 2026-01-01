package statemachine

import (
	"github.com/zellydev-games/opensplit/bridge"
	"github.com/zellydev-games/opensplit/dispatcher"
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

func (n *NewFile) ID() StateID {
	return NEWFILE
}

func (n *NewFile) OnEnter() error {
	bridge.EmitUIEvent(machine.runtimeProvider, bridge.AppViewModel{
		View:               bridge.AppViewNewSplitFile,
		SpeedrunAPIBaseURL: machine.configService.SpeedRunAPIBase,
	})
	return nil
}
func (n *NewFile) OnExit() error { return nil }
func (n *NewFile) Receive(command dispatcher.Command, payload *string) (dispatcher.DispatchReply, error) {
	switch command {
	case dispatcher.CANCEL:
		machine.changeState(WELCOME)
	case dispatcher.SUBMIT:
		if payload == nil {
			return dispatcher.DispatchReply{
				Code:    1,
				Message: "nil payload received",
			}, nil
		}
		dto, err := adapters.JSONSplitFileToDTO(*payload)
		if err != nil {
			logger.Error(err.Error())
			return dispatcher.DispatchReply{Code: 2, Message: err.Error()}, err
		}
		err = machine.repoService.SaveSplitFile(dto, 100, 100, 390, 550)
		if err != nil {
			return dispatcher.DispatchReply{Code: 4, Message: "failed to save dto: " + err.Error()}, err
		}
		sf, err := adapters.DTOSplitFileToDomain(dto)
		if err != nil {
			return dispatcher.DispatchReply{Code: 5, Message: err.Error()}, err
		}
		machine.sessionService.SetLoadedSplitFile(sf)
		machine.changeState(RUNNING)
		return dispatcher.DispatchReply{}, nil
	default:
		panic("unhandled default case")
	}
	return dispatcher.DispatchReply{}, nil
}
