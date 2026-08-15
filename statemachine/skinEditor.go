package statemachine

import (
	"encoding/json"
	"errors"

	"github.com/zellydev-games/opensplit/bridge"
	"github.com/zellydev-games/opensplit/command"
	"github.com/zellydev-games/opensplit/dispatcher"
	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/logger"
)

type SkinEditor struct {
	initialLayout string
}

func NewSkinEditorState(layout string) (*SkinEditor, error) {
	if layout != "horizontal" && layout != "vertical" {
		layout = "vertical"
	}

	return &SkinEditor{
		initialLayout: layout,
	}, nil
}

func (s *SkinEditor) OnEnter() error {
	logger.Debugf(
		logModule,
		"opening skin editor with layout %q",
		s.initialLayout,
	)

	return machine.skinProvider.EmitSkinModel()
}

func (s *SkinEditor) EmitUI() error {
	bridge.EmitUIEvent(
		machine.runtimeProvider,
		bridge.AppViewModel{
			View: bridge.AppViewSkinEditor,
		},
	)

	return nil
}

func (s *SkinEditor) OnExit() error {
	return nil
}

func (s *SkinEditor) Receive(
	c command.Command,
	payload *string,
) (
	dispatcher.DispatchReply,
	error,
) {
	switch c {
	case command.SKIN_FILE:
		return s.selectFile(payload)

	case command.SKIN_ELEMENT:
		return s.selectElement(payload)

	case command.CLEAR_ELEMENT:
		return s.clearElement()

	case command.SKIN_RULE:
		return s.selectRule(payload)

	case command.SKIN_RULE_UPDATE:
		return s.updateRule(payload)

	case command.SKIN_CREATE_FILE:
		return s.createFile(payload)

	case command.SKIN_FILE_UPDATE:
		return s.updateFile(payload)

	case command.SKIN_CREATE_RULE:
		return s.createRule(payload)

	case command.SKIN_RULE_DELETE:
		return s.deleteRule(payload)

	case command.SKIN_PREVIEW_SET:
		return s.setPreviewElements(payload)

	case command.SKIN_SET_DEFAULT_LAYOUT:
		return s.setDefaultLayout(payload)

	case command.SKIN_SAVE:
		return s.save()

	case command.SKIN_RELOAD:
		return s.reload()

	case command.CANCEL:
		return s.cancel()
	default:
		return dispatcher.DispatchReply{}, nil
	}
}

func (s *SkinEditor) cancel() (dispatcher.DispatchReply, error) {
	machine.changeState(EDITSKIN)

	return dispatcher.DispatchReply{}, nil
}

func (s *SkinEditor) setDefaultLayout(
	payload *string,
) (dispatcher.DispatchReply, error) {
	var request struct {
		Layout string `json:"layout"`
	}

	if err := decodePayload(payload, &request); err != nil {
		return errorReply(err)
	}

	if request.Layout != "horizontal" && request.Layout != "vertical" {
		return errorReply(
			errors.New("invalid skin layout"),
		)
	}

	if err := machine.skinProvider.SetDefaultLayout(request.Layout); err != nil {
		return errorReply(err)
	}

	return successReply("skin default layout updated")
}

func (s *SkinEditor) selectFile(payload *string) (dispatcher.DispatchReply, error) {
	var request struct {
		File string `json:"file"`
	}

	if err := decodePayload(payload, &request); err != nil {
		return errorReply(err)
	}

	if err := machine.skinProvider.SelectFile(request.File); err != nil {
		return errorReply(err)
	}

	return successReply("file selected")
}

func (s *SkinEditor) selectElement(payload *string) (dispatcher.DispatchReply, error) {
	var request struct {
		Element string `json:"element"`
	}

	if err := decodePayload(payload, &request); err != nil {
		return errorReply(err)
	}

	if err := machine.skinProvider.SelectElement(request.Element); err != nil {
		return errorReply(err)
	}

	return successReply("element selected")

}

func (s *SkinEditor) clearElement() (dispatcher.DispatchReply, error) {
	if err := machine.skinProvider.ClearElement(); err != nil {
		return errorReply(err)
	}

	return successReply("element cleared")

}

func (s *SkinEditor) selectRule(payload *string) (dispatcher.DispatchReply, error) {
	var request struct {
		File   string `json:"file"`
		RuleID string `json:"ruleId"`
	}

	if err := decodePayload(payload, &request); err != nil {
		return errorReply(err)
	}

	if err := machine.skinProvider.SelectRule(
		request.File,
		request.RuleID,
	); err != nil {
		return errorReply(err)
	}

	return successReply("rule selected")

}

func (s *SkinEditor) updateRule(payload *string) (dispatcher.DispatchReply, error) {
	var request struct {
		Rule dto.CSSRuleEditor `json:"rule"`
	}

	if err := decodePayload(payload, &request); err != nil {
		return errorReply(err)
	}

	if err := machine.skinProvider.UpdateActiveRule(request.Rule); err != nil {
		return errorReply(err)
	}

	return successReply("rule updated")

}

func (s *SkinEditor) createFile(payload *string) (dispatcher.DispatchReply, error) {
	var request struct {
		Name string `json:"name"`
	}

	if err := decodePayload(payload, &request); err != nil {
		return errorReply(err)
	}

	if err := machine.skinProvider.CreateCSSFile(request.Name); err != nil {
		return errorReply(err)
	}

	return successReply("css file created")
}

func (s *SkinEditor) updateFile(payload *string) (dispatcher.DispatchReply, error) {
	var request struct {
		File     string `json:"file"`
		Contents string `json:"contents"`
	}

	if err := decodePayload(payload, &request); err != nil {
		return errorReply(err)
	}

	if err := machine.skinProvider.UpdateFileContents(
		request.File,
		request.Contents,
	); err != nil {
		return errorReply(err)
	}

	return successReply("file updated")

}

func (s *SkinEditor) createRule(payload *string) (dispatcher.DispatchReply, error) {
	var request struct {
		Rule dto.CSSRuleEditor `json:"rule"`
	}

	if err := decodePayload(payload, &request); err != nil {
		return errorReply(err)
	}

	logger.Infof(
		logModule,
		"creating CSS rule: file=%q selector=%q parent=%q",
		request.Rule.File,
		request.Rule.Selector,
		request.Rule.ParentID,
	)

	if err := machine.skinProvider.CreateCSSRule(request.Rule); err != nil {
		return errorReply(err)
	}

	return successReply("css rule created")
}

func (s *SkinEditor) deleteRule(payload *string) (dispatcher.DispatchReply, error) {
	var request struct {
		RuleID string `json:"ruleId"`
	}

	if err := decodePayload(payload, &request); err != nil {
		return errorReply(err)
	}

	if err := machine.skinProvider.DeleteCSSRule(request.RuleID); err != nil {
		return errorReply(err)
	}

	return successReply("css rule deleted")
}

func (s *SkinEditor) setPreviewElements(payload *string) (dispatcher.DispatchReply, error) {
	var request []dto.SkinPreviewElement

	if err := decodePayload(payload, &request); err != nil {
		return errorReply(err)
	}

	if err := machine.skinProvider.SetPreviewElements(request); err != nil {
		return errorReply(err)
	}

	return successReply("preview elements updated")
}

func (s *SkinEditor) save() (dispatcher.DispatchReply, error) {
	logger.Infof(
		logModule,
		"saving skin working copy",
	)

	if err := machine.skinProvider.SaveWorkingCopy(); err != nil {
		return errorReply(err)
	}

	return successReply("skin saved")
}

func (s *SkinEditor) reload() (dispatcher.DispatchReply, error) {
	if err := machine.skinProvider.ReloadEditor(); err != nil {
		return errorReply(err)
	}

	if err := machine.skinProvider.EmitSkinModel(); err != nil {
		return errorReply(err)
	}

	return successReply("editor reloaded")
}

func (s *SkinEditor) String() string {
	return "Skin Editor"
}

func (s *SkinEditor) ID() StateID {
	return SKINEDITOR
}

func decodePayload(payload *string, target any) error {
	if payload == nil {
		return errors.New("command requires a payload")
	}

	return json.Unmarshal([]byte(*payload), target)
}

func successReply(
	message string,
) (
	dispatcher.DispatchReply,
	error,
) {
	return dispatcher.DispatchReply{
		Message: message,
	}, nil
}

func errorReply(
	err error,
) (
	dispatcher.DispatchReply,
	error,
) {
	return dispatcher.DispatchReply{
			Code:    1,
			Message: err.Error(),
		},
		err
}
