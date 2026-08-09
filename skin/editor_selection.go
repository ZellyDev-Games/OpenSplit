package skin

import (
	"fmt"

	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/skin/parser"
)

// SelectFile selects a CSS file in the editor.
//
// Selecting a file clears the active rule and puts the editor into file
// mode. An empty file clears the selection entirely.
func (s *Service) SelectFile(
	file string,
) error {
	if file == "" {
		s.clearFileSelection()

		return s.EmitSkinModel()
	}

	if !s.cssFileExists(file) {
		return fmt.Errorf(
			"unknown css file %q",
			file,
		)
	}

	target := s.editor.GetTarget()

	target.File = file
	target.RuleID = ""
	target.ParentID = ""
	target.Mode = "file"

	s.editor.SetTarget(target)
	s.editor.SetActiveRule(nil)

	return s.EmitSkinModel()
}

// SelectElement selects a preview element.
//
// Element selection establishes the element's selector as the target and
// searches the complete parsed rule tree for an existing matching rule.
//
// If a matching rule is found, the editor enters "existing" mode.
// Otherwise it remains in "create" mode so the frontend can create a new
// rule for the selected element.
func (s *Service) SelectElement(
	id string,
) error {
	selected, err := s.findSkinElement(id)
	if err != nil {
		return err
	}

	_, rules, _, _, _, _ := s.editor.Snapshot()

	target := dto.SkinEditorTarget{
		ElementID: id,
		Selector:  selected.Selector,
		Mode:      "create",
	}

	active := s.findMatchingRuleEditor(
		rules,
		selected.Selector,
		&target,
	)

	s.editor.SetTarget(target)
	s.editor.SetActiveRule(active)

	return s.EmitSkinModel()
}

// SelectRule selects a specific parsed CSS rule.
func (s *Service) SelectRule(
	file string,
	ruleID string,
) error {
	_, rules, _, _, _, _ := s.editor.Snapshot()

	selected, ok := findRule(
		rules,
		file,
		ruleID,
	)

	if !ok {
		return fmt.Errorf(
			"rule %s not found",
			ruleID,
		)
	}

	target := s.editor.GetTarget()

	target.File = file
	target.RuleID = ruleID
	target.ParentID = selected.ParentID
	target.Selector = selected.Selector
	target.Mode = "existing"

	s.editor.SetTarget(target)
	s.editor.SetActiveRule(
		newCSSRuleEditor(selected),
	)

	return s.EmitSkinModel()
}

// UpdateActiveRule updates the declarations of the currently selected rule.
//
// This method is used both for existing rules and for rules that were just
// created during the current editor session. A newly created rule already
// exists in the working rule tree, so its declarations can be updated in
// exactly the same way as an existing rule.
//
// The parser rule receives the parsed declarations while the active editor
// rule retains the exact body text supplied by the frontend. This keeps the
// textarea stable while the user is typing while still giving the CSS writer
// structured declarations to persist.
func (s *Service) UpdateActiveRule(
	updated dto.CSSRuleEditor,
) error {
	if err := validateActiveRuleUpdate(updated); err != nil {
		return err
	}

	declarations := parseDeclarations(
		updated.Body,
	)

	if !s.editor.UpdateRuleRecursive(
		updated.ID,
		declarations,
	) {
		return fmt.Errorf(
			"rule %s not found",
			updated.ID,
		)
	}

	// Preserve the exact text entered by the user in the active editor.
	//
	// The parser representation above contains the structured version used
	// by the CSS writer. ActiveRule is UI state and should not unexpectedly
	// reformat the textarea on every keystroke.
	s.editor.SetActiveRule(
		&updated,
	)

	return s.EmitSkinModel()
}

// clearFileSelection resets the editor selection to file mode without
// selecting a specific file.
func (s *Service) clearFileSelection() {
	target := s.editor.GetTarget()

	target.File = ""
	target.RuleID = ""
	target.ParentID = ""
	target.Mode = "file"

	s.editor.SetTarget(target)
	s.editor.SetActiveRule(nil)
}

// cssFileExists reports whether the working copy contains the requested
// CSS file.
func (s *Service) cssFileExists(
	file string,
) bool {
	files, _, _, _, _, _ := s.editor.Snapshot()

	for _, candidate := range files {
		if candidate.Path == file {
			return true
		}
	}

	return false
}

// findMatchingRuleEditor searches the complete rule tree for a concrete
// selector matching the requested element selector.
//
// When a match is found, target is updated with the matching rule's file,
// ID, parent ID, and existing mode.
func (s *Service) findMatchingRuleEditor(
	rules []parser.Rule,
	selector string,
	target *dto.SkinEditorTarget,
) *dto.CSSRuleEditor {
	if target == nil {
		return nil
	}

	var active *dto.CSSRuleEditor

	walkRules(
		rules,
		func(rule parser.Rule) bool {
			// Only concrete selector rules can directly style an
			// element. At-rules such as @media, @supports, and
			// @layer are represented by their children.
			if rule.Selector == "" {
				return false
			}

			if !parser.SelectorMatch(
				rule.Selector,
				selector,
			) {
				return false
			}

			target.File = rule.File
			target.RuleID = rule.ID
			target.ParentID = rule.ParentID
			target.Mode = "existing"

			active = newCSSRuleEditor(&rule)

			return true
		},
	)

	return active
}

// validateActiveRuleUpdate validates the editor state before a rule is
// modified.
//
// Both existing rules and rules created during the current editor session
// are valid here. Create only describes how the rule originally entered the
// working copy; it does not prevent the rule from subsequently receiving
// declaration updates.
func validateActiveRuleUpdate(
	updated dto.CSSRuleEditor,
) error {
	if updated.ID == "" {
		return fmt.Errorf(
			"cannot update rule without an id",
		)
	}

	return nil
}
