package skin

import (
	"fmt"
	"path/filepath"

	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/skin/editor"
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

	selected, ok := editor.FindRule(
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
		editor.NewCSSRuleEditor(selected),
	)

	return s.EmitSkinModel()
}

// UpdateActiveRule updates the declarations of the currently selected rule.
//
// Both existing rules and rules created during the current editor session
// are valid here.
//
// The parser rule is updated first. The complete CSS file is then regenerated
// into CSSFile.Contents so the file working copy and rule working copy remain
// synchronized.
//
// CSSFile.Contents is the representation that SaveWorkingCopy persists to
// disk.
func (s *Service) UpdateActiveRule(
	updated dto.CSSRuleEditor,
) error {
	if err := validateActiveRuleUpdate(updated); err != nil {
		return err
	}

	declarations := editor.ParseDeclarations(
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

	files,
		rules,
		_,
		_,
		_,
		_ := s.editor.Snapshot()

	if err := synchronizeCSSFileContents(
		files,
		rules,
		updated.File,
	); err != nil {
		return err
	}

	s.replaceWorkingCopy(
		files,
		rules,
	)

	// Preserve the exact text entered by the user in the active editor.
	//
	// The parser representation contains the structured declarations used by
	// the CSS writer, while ActiveRule is UI state.
	s.editor.SetActiveRule(
		&updated,
	)

	return s.EmitSkinModel()
}

// UpdateFileContents updates the contents of one working-copy file.
//
// CSS files are reparsed immediately so that direct text edits and structured
// rule editing continue to operate on the same working copy.
//
// Non-CSS text files only require their file contents to be updated.
func (s *Service) UpdateFileContents(
	file string,
	contents string,
) error {
	files,
		rules,
		_,
		_,
		_,
		_ := s.editor.Snapshot()

	found := false

	for i := range files {
		if files[i].Path != file {
			continue
		}

		files[i].Contents = contents
		found = true

		break
	}

	if !found {
		return fmt.Errorf(
			"file %s not found",
			file,
		)
	}

	if filepath.Ext(file) == ".css" {
		updatedRules := parseCSSFileContents(
			file,
			contents,
		)

		rules = replaceRulesForFile(
			rules,
			file,
			updatedRules,
		)

		// Direct file editing invalidates the previously selected parsed
		// rule because parser IDs may have changed.
		target := s.editor.GetTarget()

		if target.File == file {
			target.RuleID = ""
			target.ParentID = ""
			target.Mode = "file"

			s.editor.SetTarget(target)
			s.editor.SetActiveRule(nil)
		}
	}

	s.replaceWorkingCopy(
		files,
		rules,
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

	editor.WalkRules(
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

			active = editor.NewCSSRuleEditor(&rule)

			return true
		},
	)

	return active
}

// findSkinElement returns the editor element with the requested ID.
func (s *Service) findSkinElement(
	id string,
) (*dto.SkinElement, error) {
	if id == "" {
		return nil, fmt.Errorf(
			"element id cannot be empty",
		)
	}

	elements := s.GetSkinElements()

	for _, element := range elements {
		if element.ID != id {
			continue
		}

		copy := element

		return &copy, nil
	}

	return nil, fmt.Errorf(
		"unknown element %q",
		id,
	)
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

// parseCSSFileContents parses one CSS file into the editor rule tree.
func parseCSSFileContents(
	file string,
	contents string,
) []parser.Rule {
	if filepath.Ext(file) != ".css" {
		return nil
	}

	return parser.Parse(
		file,
		contents,
	)
}

// replaceRulesForFile replaces the top-level rules belonging to one CSS file.
//
// Rules belonging to other files remain untouched.
func replaceRulesForFile(
	rules []parser.Rule,
	file string,
	replacement []parser.Rule,
) []parser.Rule {
	out := make(
		[]parser.Rule,
		0,
		len(rules),
	)

	inserted := false

	for _, rule := range rules {
		if rule.File == file {
			if !inserted {
				out = append(
					out,
					replacement...,
				)

				inserted = true
			}

			continue
		}

		out = append(
			out,
			rule,
		)
	}

	if !inserted {
		out = append(
			out,
			replacement...,
		)
	}

	return out
}

// synchronizeCSSFileContents regenerates the working-copy contents for one
// CSS file from its current parser rule tree.
//
// This is used after a structured rule edit so the raw file representation
// remains synchronized with the parsed rule representation.
func synchronizeCSSFileContents(
	files []dto.CSSFile,
	rules []parser.Rule,
	file string,
) error {
	if filepath.Ext(file) != ".css" {
		return nil
	}

	fileRules := make(
		[]parser.Rule,
		0,
	)

	for _, rule := range rules {
		if rule.File != file {
			continue
		}

		fileRules = append(
			fileRules,
			rule,
		)
	}

	contents := parser.FormatCSS(
		fileRules,
	)

	for i := range files {
		if files[i].Path != file {
			continue
		}

		files[i].Contents = contents

		return nil
	}

	return fmt.Errorf(
		"file %s not found",
		file,
	)
}
