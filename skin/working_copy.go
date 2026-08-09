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

// package skin

// import (
// 	"errors"
// 	"fmt"

// 	"github.com/zellydev-games/opensplit/dto"
// 	"github.com/zellydev-games/opensplit/logger"
// 	"github.com/zellydev-games/opensplit/skin/parser"
// )

// ClearElement clears the currently selected editor element.
func (s *Service) ClearElement() error {
	s.editor.ClearSelection()

	return s.EmitSkinModel()
}

// // CreateCSSRule adds a new CSS rule to the editor working copy.
// //
// // New rules are inserted into the requested parent rule when ParentID is
// // provided. This preserves the parser rule tree used by the CSS writer.
// func (s *Service) CreateCSSRule(
// 	rule dto.CSSRuleEditor,
// ) error {
// 	if rule.File == "" {
// 		return errors.New(
// 			"css rule requires file",
// 		)
// 	}

// 	if rule.Selector == "" {
// 		return errors.New(
// 			"css rule requires selector",
// 		)
// 	}

// 	files,
// 		rules,
// 		_,
// 		_,
// 		_,
// 		_ := s.editor.Snapshot()

// 	editorRule, cssRule := buildCSSRule(
// 		rule,
// 	)

// 	var inserted bool

// 	rules, inserted = insertRuleRecursive(
// 		rules,
// 		rule.ParentID,
// 		cssRule,
// 	)

// 	logger.Infof(
// 		logModule,
// 		"CreateCSSRule: inserted=%t id=%q file=%q selector=%q parent=%q",
// 		inserted,
// 		editorRule.ID,
// 		editorRule.File,
// 		editorRule.Selector,
// 		editorRule.ParentID,
// 	)

// 	if !inserted {
// 		return fmt.Errorf(
// 			"parent rule %s not found",
// 			rule.ParentID,
// 		)
// 	}

// 	s.replaceWorkingCopy(
// 		files,
// 		rules,
// 	)

// 	_, rulesAfter, dirtyAfter, _, _, revisionAfter := s.editor.Snapshot()

// 	logger.Infof(
// 		logModule,
// 		"CreateCSSRule: working copy after mutation rules=%d dirty=%t revision=%d",
// 		len(rulesAfter),
// 		dirtyAfter,
// 		revisionAfter,
// 	)

// 	s.editor.SetActiveRule(
// 		&editorRule,
// 	)

// 	target := s.editor.GetTarget()

// 	target.File = editorRule.File
// 	target.RuleID = editorRule.ID
// 	target.Selector = editorRule.Selector
// 	target.Mode = "existing"

// 	s.editor.SetTarget(
// 		target,
// 	)

// 	return s.EmitSkinModel()
// }

// DeleteCSSRule removes a CSS rule from the working copy.
//
// Rules nested inside at-rules are supported.
func (s *Service) DeleteCSSRule(
	id string,
) error {
	if !s.editor.DeleteRuleRecursive(id) {
		return fmt.Errorf(
			"rule %s not found",
			id,
		)
	}

	s.editor.SetActiveRule(nil)

	return s.EmitSkinModel()
}

// UpdateFileContents updates the contents of one CSS file in the working
// copy.
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

	s.replaceWorkingCopy(
		files,
		rules,
	)

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

// // buildCSSRule converts an editor rule into the parser representation used by
// // the working copy while preserving the editor representation for the
// // frontend.
// func buildCSSRule(
// 	rule dto.CSSRuleEditor,
// ) (dto.CSSRuleEditor, parser.Rule) {
// 	id := rule.ID

// 	if id == "" {
// 		id = fmt.Sprintf(
// 			"%s:%s",
// 			rule.File,
// 			rule.Selector,
// 		)
// 	}

// 	editorRule := rule
// 	editorRule.ID = id

// 	declarations := parseDeclarations(
// 		rule.Body,
// 	)

// 	cssRule := parser.Rule{
// 		ID:           id,
// 		File:         rule.File,
// 		Layer:        rule.Layer,
// 		Selector:     rule.Selector,
// 		ParentID:     rule.ParentID,
// 		Declarations: declarations,
// 	}

// 	return editorRule, cssRule
// }

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
