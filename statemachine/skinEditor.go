package statemachine

import (
	"encoding/json"

	"github.com/zellydev-games/opensplit/bridge"
	"github.com/zellydev-games/opensplit/command"
	"github.com/zellydev-games/opensplit/dispatcher"
	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/logger"
)

type SkinEditor struct {
}

func NewSkinEditorState() (*SkinEditor, error) {
	return &SkinEditor{}, nil
}

func (s *SkinEditor) OnEnter() error {

	logger.Debug(
		logModule,
		"opening skin editor",
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

	case command.CANCEL:

		machine.changeState(
			EDITSKIN,
		)

		return dispatcher.DispatchReply{}, nil

	case command.SKIN_FILE:

		var request struct {
			File string `json:"file"`
		}

		if err := json.Unmarshal(
			[]byte(*payload),
			&request,
		); err != nil {

			return errorReply(err)
		}

		if err := machine.skinProvider.SelectFile(
			request.File,
		); err != nil {

			return errorReply(err)
		}

		return successReply(
			"file selected",
		)

	case command.SKIN_ELEMENT:

		var request struct {
			Element string `json:"element"`
		}

		if err := json.Unmarshal(
			[]byte(*payload),
			&request,
		); err != nil {

			return errorReply(err)
		}

		if err := machine.skinProvider.SelectElement(
			request.Element,
		); err != nil {

			return errorReply(err)
		}

		return successReply(
			"element selected",
		)

	case command.CLEAR_ELEMENT:

		if err := machine.skinProvider.ClearElement(); err != nil {
			return errorReply(err)
		}

		return successReply(
			"element cleared",
		)

	case command.SKIN_RULE:

		var request struct {
			File   string `json:"file"`
			RuleID string `json:"ruleId"`
		}

		if err := json.Unmarshal(
			[]byte(*payload),
			&request,
		); err != nil {

			return errorReply(err)
		}

		if err := machine.skinProvider.SelectRule(
			request.File,
			request.RuleID,
		); err != nil {

			return errorReply(err)
		}

		return successReply(
			"rule selected",
		)

	case command.SKIN_RULE_UPDATE:

		var request struct {
			Rule dto.CSSRuleEditor `json:"rule"`
		}

		if err := json.Unmarshal(
			[]byte(*payload),
			&request,
		); err != nil {

			return errorReply(err)
		}

		if err := machine.skinProvider.UpdateActiveRule(
			request.Rule,
		); err != nil {

			return errorReply(err)
		}

		return successReply(
			"rule updated",
		)

	case command.SKIN_CREATE_FILE:

		var request struct {
			Name string `json:"name"`
		}

		if err := json.Unmarshal(
			[]byte(*payload),
			&request,
		); err != nil {

			return errorReply(err)
		}

		if err := machine.skinProvider.CreateCSSFile(
			request.Name,
		); err != nil {

			return errorReply(err)
		}

		return successReply(
			"css file created",
		)

	case command.SKIN_FILE_UPDATE:

		var request struct {
			File     string `json:"file"`
			Contents string `json:"contents"`
		}

		if err := json.Unmarshal(
			[]byte(*payload),
			&request,
		); err != nil {

			return errorReply(err)
		}

		if err := machine.skinProvider.UpdateFileContents(
			request.File,
			request.Contents,
		); err != nil {

			return errorReply(err)
		}

		return successReply(
			"file updated",
		)

	case command.SKIN_CREATE_RULE:

		var request struct {
			Rule dto.CSSRuleEditor `json:"rule"`
		}

		if err := json.Unmarshal(
			[]byte(*payload),
			&request,
		); err != nil {

			return errorReply(err)
		}

		if err := machine.skinProvider.CreateCSSRule(
			request.Rule,
		); err != nil {

			return errorReply(err)
		}

		return successReply(
			"css rule created",
		)

	case command.SKIN_RULE_DELETE:

		var request struct {
			RuleID string `json:"ruleId"`
		}

		if err := json.Unmarshal(
			[]byte(*payload),
			&request,
		); err != nil {

			return errorReply(err)
		}

		if err := machine.skinProvider.DeleteCSSRule(
			request.RuleID,
		); err != nil {

			return errorReply(err)
		}

		return successReply(
			"css rule deleted",
		)

	case command.SKIN_PREVIEW_SET:

		var request []dto.SkinPreviewElement

		if err := json.Unmarshal(
			[]byte(*payload),
			&request,
		); err != nil {
			return errorReply(err)
		}

		if err := machine.skinProvider.SetPreviewElements(request); err != nil {
			return errorReply(err)
		}

		return successReply("preview elements updated")

	case command.SKIN_SAVE:

		if err := machine.skinProvider.SaveWorkingCopy(); err != nil {
			return errorReply(err)
		}

		return successReply(
			"skin saved",
		)

	case command.SKIN_RELOAD:

		if err := machine.skinProvider.ReloadEditor(); err != nil {
			return errorReply(err)
		}

		if err := machine.skinProvider.EmitSkinModel(); err != nil {
			return errorReply(err)
		}

		return successReply(
			"editor reloaded",
		)

	default:

		return dispatcher.DispatchReply{}, nil
	}
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

func (s *SkinEditor) String() string {
	return "Skin Editor"
}

func (s *SkinEditor) ID() StateID {
	return SKINEDITOR
}
