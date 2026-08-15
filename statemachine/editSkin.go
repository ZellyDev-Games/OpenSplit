package statemachine

import (
	"encoding/json"

	"github.com/zellydev-games/opensplit/bridge"
	"github.com/zellydev-games/opensplit/command"
	"github.com/zellydev-games/opensplit/dispatcher"
	"github.com/zellydev-games/opensplit/logger"
)

// EditSkin controls selecting and editing an existing skin.
type EditSkin struct {
}

func NewEditSkinState() (*EditSkin, error) {
	return &EditSkin{}, nil
}

// OnEnter displays the skin selection screen.
// The editor is not opened until a skin has been selected.
func (s *EditSkin) OnEnter() error {
	return machine.skinProvider.EmitSkinModel()
}

func (s *EditSkin) EmitUI() error {
	bridge.EmitUIEvent(
		machine.runtimeProvider,
		bridge.AppViewModel{
			View: bridge.AppViewEditSkin,
		},
	)

	return nil
}

func (s *EditSkin) OnExit() error {
	return nil
}

func (s *EditSkin) Receive(
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

		if err := json.Unmarshal([]byte(*payload), &request); err != nil {
			return dispatcher.DispatchReply{
				Code:    2,
				Message: err.Error(),
			}, err
		}

		if request.Name == "" {
			return dispatcher.DispatchReply{
				Code:    3,
				Message: "no skin selected",
			}, nil
		}

		if request.Layout != "horizontal" && request.Layout != "vertical" {
			request.Layout = "vertical"
		}

		logger.Infof(
			logModule,
			"loading skin editor for %s with %s layout.",
			request.Name,
			request.Layout,
		)

		if err := machine.skinProvider.SetSkin(
			request.Name,
			false,
		); err != nil {
			return dispatcher.DispatchReply{
				Code:    4,
				Message: err.Error(),
			}, err
		}

		if err := machine.skinProvider.EmitSkinModel(); err != nil {
			return dispatcher.DispatchReply{
				Code:    6,
				Message: err.Error(),
			}, err
		}

		machine.changeState(
			SKINEDITOR,
			request.Layout,
		)

		return dispatcher.DispatchReply{
			Message: "skin editor opened",
		}, nil

	default:
		return dispatcher.DispatchReply{}, nil
	}
}

func (s *EditSkin) String() string {
	return "Edit Skin"
}

func (s *EditSkin) ID() StateID {
	return EDITSKIN
}
