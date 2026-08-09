package skin

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/skin/parser"
)

// CreateCSSRule adds a new CSS rule to the editor working copy.
//
// New rules are inserted into the requested parent rule when ParentID is
// provided. This preserves the parser rule tree used by the CSS writer.
func (s *Service) CreateCSSRule(
	rule dto.CSSRuleEditor,
) error {
	if rule.File == "" {
		return errors.New(
			"css rule requires file",
		)
	}

	if rule.Selector == "" {
		return errors.New(
			"css rule requires selector",
		)
	}

	files,
		rules,
		_,
		_,
		_,
		_ := s.editor.Snapshot()

	editorRule, cssRule := buildCSSRule(
		rule,
	)

	var inserted bool

	rules, inserted = insertRuleRecursive(
		rules,
		rule.ParentID,
		cssRule,
	)

	logger.Infof(
		logModule,
		"CreateCSSRule: inserted=%t id=%q file=%q selector=%q parent=%q",
		inserted,
		editorRule.ID,
		editorRule.File,
		editorRule.Selector,
		editorRule.ParentID,
	)

	if !inserted {
		return fmt.Errorf(
			"parent rule %s not found",
			rule.ParentID,
		)
	}

	// Keep the file representation synchronized with the parser rule tree.
	if err := synchronizeCSSFileContents(
		files,
		rules,
		editorRule.File,
	); err != nil {
		return err
	}

	s.replaceWorkingCopy(
		files,
		rules,
	)

	_, rulesAfter, dirtyAfter, _, _, revisionAfter := s.editor.Snapshot()

	logger.Infof(
		logModule,
		"CreateCSSRule: working copy after mutation rules=%d dirty=%t revision=%d",
		len(rulesAfter),
		dirtyAfter,
		revisionAfter,
	)

	s.editor.SetActiveRule(
		&editorRule,
	)

	target := s.editor.GetTarget()

	target.File = editorRule.File
	target.RuleID = editorRule.ID
	target.ParentID = editorRule.ParentID
	target.Selector = editorRule.Selector
	target.Mode = "existing"

	s.editor.SetTarget(
		target,
	)

	return s.EmitSkinModel()
}

// buildCSSRule converts an editor rule into the parser representation used by
// the working copy while preserving the editor representation for the
// frontend.
//
// Rules created during an editor session do not yet have a source location,
// so they cannot use the normal parser ID format of:
//
//	file:line:sibling-index
//
// They receive a unique working-copy ID instead. Once the file is saved and
// reparsed, assignRuleIDs will replace that temporary identity with the
// source-based parser ID.
func buildCSSRule(
	rule dto.CSSRuleEditor,
) (dto.CSSRuleEditor, parser.Rule) {
	id := rule.ID

	if id == "" {
		id = fmt.Sprintf(
			"%s:new:%s",
			rule.File,
			uuid.NewString(),
		)
	}

	editorRule := rule
	editorRule.ID = id

	declarations := parseDeclarations(
		rule.Body,
	)

	cssRule := parser.Rule{
		ID:           id,
		File:         rule.File,
		Layer:        rule.Layer,
		Selector:     rule.Selector,
		ParentID:     rule.ParentID,
		Declarations: declarations,
	}

	return editorRule, cssRule
}

// ClearElement clears the currently selected editor element.
func (s *Service) ClearElement() error {
	s.editor.ClearSelection()

	return s.EmitSkinModel()
}

// DeleteCSSRule removes a CSS rule from the working copy.
//
// Rules nested inside at-rules are supported.
func (s *Service) DeleteCSSRule(
	id string,
) error {
	files,
		rules,
		_,
		_,
		_,
		_ := s.editor.Snapshot()

	selected, ok := findRuleByID(
		rules,
		id,
	)

	if !ok {
		return fmt.Errorf(
			"rule %s not found",
			id,
		)
	}

	if !s.editor.DeleteRuleRecursive(id) {
		return fmt.Errorf(
			"rule %s not found",
			id,
		)
	}

	_, updatedRules, _, _, _, _ := s.editor.Snapshot()

	if err := synchronizeCSSFileContents(
		files,
		updatedRules,
		selected.File,
	); err != nil {
		return err
	}

	s.replaceWorkingCopy(
		files,
		updatedRules,
	)

	s.editor.SetActiveRule(nil)

	return s.EmitSkinModel()
}

// replaceWorkingCopy is the common mutation boundary for editor working-copy
// changes.
func (s *Service) replaceWorkingCopy(
	files []dto.CSSFile,
	rules []parser.Rule,
) {
	s.editor.ReplaceWorkingCopy(
		files,
		rules,
	)
}

// insertRuleRecursive inserts rule into the requested parent.
//
// An empty parent ID means the rule belongs at the root of the rule tree.
func insertRuleRecursive(
	rules []parser.Rule,
	parentID string,
	rule parser.Rule,
) ([]parser.Rule, bool) {
	if parentID == "" {
		return append(
			rules,
			rule,
		), true
	}

	for i := range rules {
		if rules[i].ID == parentID {
			rules[i].Children = append(
				rules[i].Children,
				rule,
			)

			return rules, true
		}

		updated, inserted := insertRuleRecursive(
			rules[i].Children,
			parentID,
			rule,
		)

		if inserted {
			rules[i].Children = updated

			return rules, true
		}
	}

	return rules, false
}
