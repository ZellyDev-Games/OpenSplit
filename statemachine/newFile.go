package statemachine

import (
	"github.com/zellydev-games/opensplit/bridge"
	"github.com/zellydev-games/opensplit/command"
	"github.com/zellydev-games/opensplit/dispatcher"
	"github.com/zellydev-games/opensplit/dto"
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
		View: bridge.AppViewNewSplitFile,
		SplitFile: &dto.SplitFile{
			SelectedSkin: machine.configService.SelectedSkin,
		},
	})

	return nil
}

func (n *NewFile) OnExit() error { return nil }

func (n *NewFile) Receive(c command.Command, payload *string) (dispatcher.DispatchReply, error) {
	switch c {
	case command.CANCEL:
		machine.changeState(WELCOME)
	case command.SUBMIT:
		if payload == nil {
			return dispatcher.DispatchReply{
				Code:    1,
				Message: "nil payload received",
			}, nil
		}
		dto, err := adapters.JSONSplitFileToDTO(*payload)
		if err != nil {
			logger.Error(logModule, err.Error())
			return dispatcher.DispatchReply{Code: 2, Message: err.Error()}, err
		}
		err = machine.repoService.SaveSplitFile(dto)
		if err != nil {
			return dispatcher.DispatchReply{Code: 4, Message: "failed to save dto: " + err.Error()}, err
		}
		_ = machine.skinProvider.SetSkin(dto.SelectedSkin, false)
		sf, err := adapters.DTOSplitFileToDomain(dto)
		if err != nil {
			return dispatcher.DispatchReply{Code: 5, Message: err.Error()}, err
		}
		machine.sessionService.SetLoadedSplitFile(sf)
		go machine.updateWorldRecord()
		machine.changeState(RUNNING)
		return dispatcher.DispatchReply{}, nil
	default:
		panic("unhandled default case")
	}
	return dispatcher.DispatchReply{}, nil
}
