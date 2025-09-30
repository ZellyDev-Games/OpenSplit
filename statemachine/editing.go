package statemachine

import (
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
	payload := machine.sessionService.SplitFile()
	machine.sessionService.Pause()
	machine.runtimeProvider.EventsEmit("state:enter", EDITING, payload)
	return nil
}

func (e *Editing) OnExit() error { return nil }
func (e *Editing) Receive(command dispatcher.Command, payload *string) (dispatcher.DispatchReply, error) {
	switch command {
	case dispatcher.CANCEL:
		machine.changeState(RUNNING)
	case dispatcher.SUBMIT:
		if payload == nil {
			return dispatcher.DispatchReply{
				Code:    1,
				Message: "nil payload received",
			}, nil
		}
		dto, err := adapters.FrontendToSplitFile(*payload)
		if err != nil {
			return dispatcher.DispatchReply{2, err.Error()}, err
		}
		err = machine.repoService.SaveSplitFile(dto, dto.WindowX, dto.WindowY, dto.WindowWidth, dto.WindowHeight)
		if err != nil {
			return dispatcher.DispatchReply{4, "failed to save dto: " + err.Error()}, err
		}

		sf, err := adapters.SplitFileToDomain(dto)
		if err != nil {
			return dispatcher.DispatchReply{5, err.Error()}, err
		}
		machine.sessionService.SetLoadedSplitFile(sf)
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
