package skin

import (
	"fmt"

	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/skin/parser"
)

func (s *Service) SelectFile(
	file string,
) error {

	if file == "" {
		target := s.editor.GetTarget()

		target.File = ""
		target.RuleID = ""
		target.Mode = "file"

		s.editor.SetTarget(target)

		s.editor.SetActiveRule(nil)

		return s.EmitSkinModel()
	}

	files,
		_,
		_,
		_,
		_,
		_ :=
		s.editor.Snapshot()

	for _, f := range files {

		if f.Path == file {

			target := s.editor.GetTarget()

			target.File = file
			target.Mode = "file"
			target.RuleID = ""

			s.editor.SetTarget(
				target,
			)

			s.editor.SetActiveRule(
				nil,
			)

			return s.EmitSkinModel()
		}
	}

	return fmt.Errorf(
		"unknown css file %q",
		file,
	)
}

func (s *Service) SelectElement(
	id string,
) error {

	elements :=
		s.GetSkinElements()

	var selected *dto.SkinElement

	for _, element := range elements {

		if element.ID == id {

			copy := element
			selected = &copy

			break
		}
	}

	if selected == nil {

		return fmt.Errorf(
			"unknown element %q",
			id,
		)
	}

	_,
		rules,
		_,
		_,
		_,
		_ :=
		s.editor.Snapshot()

	target :=
		dto.SkinEditorTarget{

			ElementID: id,

			Selector: selected.Selector,

			Mode: "create",
		}

	var active *dto.CSSRuleEditor

	walkRules(
		rules,
		func(rule parser.Rule) bool {
			// Ignore:
			//
			// @media
			// @supports
			// @font-face
			// @import
			//
			// They do not directly style elements.
			// if len(rule.Selectors) == 0 {
			// 	return false
			// }
			if rule.Selector == "" {
				return false
			}

			if !parser.SelectorMatch(
				rule.Selector,
				selected.Selector,
			) {
				return false
			}

			target.File =
				rule.File

			target.RuleID =
				rule.ID

			target.Mode =
				"existing"

			active =
				&dto.CSSRuleEditor{

					ID: rule.ID,

					File: rule.File,

					Selector: rule.Selector,

					Layer: rule.Layer,

					Body: formatRuleDeclarations(
						rule.Declarations,
					),

					OriginalFile: rule.File,

					OriginalID: rule.ID,
				}

			return true
		},
	)

	s.editor.SetTarget(
		target,
	)

	s.editor.SetActiveRule(
		active,
	)

	return s.EmitSkinModel()
}

func (s *Service) SelectRule(
	file string,
	ruleID string,
) error {

	_,
		rules,
		_,
		_,
		_,
		_ :=
		s.editor.Snapshot()

	var selected *parser.Rule

	walkRules(
		rules,
		func(rule parser.Rule) bool {

			if rule.File != file ||
				rule.ID != ruleID {

				return false
			}

			copy := rule
			selected = &copy

			return true
		},
	)

	if selected == nil {

		return fmt.Errorf(
			"rule %s not found",
			ruleID,
		)
	}

	editor :=
		dto.CSSRuleEditor{

			ID: selected.ID,

			File: selected.File,

			Layer: selected.Layer,

			Body: formatRuleDeclarations(
				selected.Declarations,
			),

			OriginalFile: selected.File,

			OriginalID: selected.ID,
		}

	editor.Selector =
		selected.Selector

	target :=
		s.editor.GetTarget()

	target.File =
		file

	target.RuleID =
		ruleID

	target.Mode =
		"existing"

	s.editor.SetTarget(
		target,
	)

	s.editor.SetActiveRule(
		&editor,
	)

	return s.EmitSkinModel()
}

func (s *Service) UpdateActiveRule(
	updated dto.CSSRuleEditor,
) error {

	files,
		rules,
		_,
		_,
		_,
		_ :=
		s.editor.Snapshot()

	if !updateRuleRecursive(
		rules,
		updated.ID,
		updated.Body,
	) {

		return fmt.Errorf(
			"rule %s not found",
			updated.ID,
		)
	}

	s.editor.ReplaceWorkingCopy(
		files,
		rules,
	)

	updated.Body =
		formatRuleDeclarations(
			parseDeclarations(
				updated.Body,
			),
		)

	s.editor.SetActiveRule(
		&updated,
	)

	return s.EmitSkinModel()
}

func updateRuleRecursive(
	rules []parser.Rule,
	id string,
	body string,
) bool {

	for i := range rules {

		if rules[i].ID == id {

			rules[i].Declarations =
				parseDeclarations(
					body,
				)

			return true
		}

		if updateRuleRecursive(
			rules[i].Children,
			id,
			body,
		) {
			return true
		}
	}

	return false
}

// parseDeclarations converts editor body text into parser declarations.
//
// This keeps the frontend editor as a simple text editor while the backend
// maintains structured CSS.
func parseDeclarations(
	body string,
) []parser.Declaration {

	rule :=
		parser.Parse(
			"editor",
			".temporary {\n"+body+"\n}",
		)

	if len(rule) == 0 {
		return nil
	}

	return rule[0].Declarations
}
