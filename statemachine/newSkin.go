package statemachine

// NewSkin creates a skin from the default template and opens the editor.

import (
	"encoding/json"

	"github.com/zellydev-games/opensplit/bridge"
	"github.com/zellydev-games/opensplit/command"
	"github.com/zellydev-games/opensplit/dispatcher"
	"github.com/zellydev-games/opensplit/logger"
)

// NewSkin creates a new editable skin from the default skin.
type NewSkin struct {
}

// NewNewSkinState creates the new skin state.
func NewNewSkinState() (*NewSkin, error) {
	return &NewSkin{}, nil
}

// OnEnter displays the new skin creation screen.
func (s *NewSkin) OnEnter() error {
	logger.Debug(logModule, "opening new skin editor")

	return nil
}

func (s *NewSkin) EmitUI() error {
	bridge.EmitUIEvent(
		machine.runtimeProvider,
		bridge.AppViewModel{
			View: bridge.AppViewNewSkin,
		},
	)

	return nil
}

func (s *NewSkin) OnExit() error {
	return nil
}

func (s *NewSkin) Receive(
	c command.Command,
	payload *string,
) (dispatcher.DispatchReply, error) {

	switch c {

	case command.CANCEL:
		machine.changeState(WELCOME)

		return dispatcher.DispatchReply{}, nil

	case command.SUBMIT:

		if payload == nil {
			return dispatcher.DispatchReply{
				Code:    1,
				Message: "nil payload received",
			}, nil
		}

		var request struct {
			Name   string `json:"name"`
			Layout string `json:"layout"`
		}

		err := json.Unmarshal(
			[]byte(*payload),
			&request,
		)

		if err != nil {
			return dispatcher.DispatchReply{
				Code:    2,
				Message: err.Error(),
			}, err
		}

		if request.Name == "" {
			return dispatcher.DispatchReply{
				Code:    3,
				Message: "skin name required",
			}, nil
		}

		if request.Layout != "horizontal" && request.Layout != "vertical" {
			request.Layout = "vertical"
		}

		logger.Infof(
			logModule,
			"creating new skin %s with layout %s.",
			request.Name,
			request.Layout,
		)

		err = machine.skinProvider.CreateSkin(
			request.Name,
			"default",
		)

		if err != nil {
			return dispatcher.DispatchReply{
				Code:    4,
				Message: err.Error(),
			}, err
		}

		err = machine.skinProvider.SetSkin(
			request.Name,
			false,
		)

		if err != nil {
			return dispatcher.DispatchReply{
				Code:    5,
				Message: err.Error(),
			}, err
		}

		if err := machine.skinProvider.EmitSkinModel(); err != nil {
			return dispatcher.DispatchReply{
				Code:    6,
				Message: err.Error(),
			}, err
		}

		// Open editor on newly created skin.
		machine.changeState(
			SKINEDITOR,
			request.Layout,
		)

		return dispatcher.DispatchReply{
			Message: "skin created",
		}, nil

	default:
		return dispatcher.DispatchReply{}, nil
	}
}

func (s *NewSkin) String() string {
	return "New Skin"
}

func (s *NewSkin) ID() StateID {
	return NEWSKIN
}
