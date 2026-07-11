package statemachine

import (
	"errors"

	"github.com/zellydev-games/opensplit/bridge"
	"github.com/zellydev-games/opensplit/command"
	"github.com/zellydev-games/opensplit/dispatcher"
	"github.com/zellydev-games/opensplit/repo/adapters"
)

// Editing indicates that the frontend should show the SplitEditor, and can also pass along a split file if loaded.
type Editing struct{}

// NewEditingState requires a *session.SplitFile as its sole payload parameter
func NewEditingState() (*Editing, error) {
	return &Editing{}, nil
}

// OnEnter sets the context from Wails, and signals the frontend to show the SplitEditor with the specified split file (or nil)
func (e *Editing) OnEnter() error {
	sf, loaded := machine.sessionService.SplitFile()
	if !loaded {
		return errors.New("editing state entered but split file not loaded")
	}

	splitFileDTO := adapters.DomainSplitFileToDTO(sf)
	machine.sessionService.Pause()
	bridge.EmitUIEvent(machine.runtimeProvider, bridge.AppViewModel{
		View:      bridge.AppViewEditSplitFile,
		SplitFile: &splitFileDTO,
	})
	return nil
}

func (e *Editing) OnExit() error { return nil }
func (e *Editing) Receive(c command.Command, payload *string) (dispatcher.DispatchReply, error) {
	switch c {
	case command.CANCEL:
		machine.changeState(RUNNING)
	case command.SUBMIT:
		if payload == nil {
			return dispatcher.DispatchReply{
				Code:    1,
				Message: "nil payload received",
			}, nil
		}
		dto, err := adapters.JSONSplitFileToDTO(*payload)
		if err != nil {
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

func (e *Editing) String() string {
	return "Editing"
}
func (e *Editing) ID() StateID {
	return EDITING
}
