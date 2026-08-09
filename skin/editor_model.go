package skin

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/skin/editor"
	"github.com/zellydev-games/opensplit/skin/parser"
)

// ReloadEditor rebuilds the editor working copy from disk.
//
// This is the only disk -> editor state path.
//
// Reload discards:
//   - created files
//   - created rules
//   - modified rules
//   - pending deletions
//
// and replaces them with the current skin files on disk.
//
// The current editor selection is preserved whenever possible. Existing rules
// are restored by their parser ID. Rules created during the current editor
// session use temporary IDs, so after a save/reload they are restored using
// their file, selector, and parent instead.
func (s *Service) ReloadEditor() error {
	logger.Infof(
		logModule,
		"ReloadEditor called",
	)

	files, err := s.buildEditorFiles()
	if err != nil {
		return err
	}

	rules := parseEditorRules(files)

	logger.Infof(
		logModule,
		"ReloadEditor loaded %d rules from disk",
		len(rules),
	)

	// Capture the complete selection before replacing the working copy.
	//
	// This is important because ReplaceWorkingCopyFromDisk replaces the
	// parsed rule tree and may invalidate temporary IDs belonging to rules
	// created during the previous editor session.
	oldTarget := s.editor.GetTarget()
	oldActiveRule := s.editor.GetActiveRule()

	s.editor.ReplaceWorkingCopyFromDisk(
		files,
		rules,
	)

	_, snapshotRules, _, _, _, _ := s.editor.Snapshot()

	logger.Infof(
		logModule,
		"snapshot contains %d rules",
		len(snapshotRules),
	)

	logEditorRules(rules)

	switch {
	case oldTarget.RuleID != "":
		if err := s.restoreRuleSelection(
			oldTarget,
			oldActiveRule,
			rules,
		); err != nil {
			logger.Infof(
				logModule,
				"unable to restore rule selection: %v",
				err,
			)

			// The selected file still exists, so preserve file selection
			// rather than clearing the entire editor selection.
			s.editor.SetTarget(
				dto.SkinEditorTarget{
					ElementID: oldTarget.ElementID,
					File:      oldTarget.File,
					Selector:  oldTarget.Selector,
					Mode:      "file",
				},
			)

			s.editor.SetActiveRule(nil)
		}

	case oldTarget.File != "":
		s.editor.SetTarget(
			dto.SkinEditorTarget{
				ElementID: oldTarget.ElementID,
				File:      oldTarget.File,
				Selector:  oldTarget.Selector,
				Mode:      "file",
			},
		)
	}

	return nil
}

// GetSkinEditorModel returns the current backend-owned editor state.
func (s *Service) GetSkinEditorModel() (
	SkinModel,
	error,
) {
	files,
		rules,
		dirty,
		target,
		active,
		revision :=
		s.editor.Snapshot()

	elements := s.GetSkinElements()

	logger.Infof(
		logModule,
		"editor elements: %d",
		len(elements),
	)

	return SkinModel{
		Name: s.SelectedSkin(),

		Skins: s.GetAvailableSkins(),

		Directory: s.GetSkinPath(),

		StyleSheet: s.GetSkinAddress(),

		Files: files,

		Elements: elements,

		Rules: editor.ConvertRules(
			rules,
		),

		Target: target,

		ActiveRule: active,

		Dirty: dirty,

		Revision: revision,
	}, nil
}

// EmitSkinModel sends the current backend-owned editor state.
//
// The model channel is buffered so a model update is not lost when the UI
// pump is between receives. If an older model is already waiting, replace it
// with the newest model because only the latest editor state is meaningful.
func (s *Service) EmitSkinModel() error {
	model, err := s.GetSkinEditorModel()
	if err != nil {
		return err
	}

	select {
	case <-s.skinModelCh:
	default:
	}

	s.skinModelCh <- model

	return nil
}

// restoreRuleSelection restores a previously selected rule after the editor
// working copy has been rebuilt.
//
// Normal parsed rules retain their parser-generated ID, so ID is the preferred
// identity.
//
// Rules created during an editor session receive temporary IDs such as:
//
//	complete.css:new:<uuid>
//
// Once those rules are saved and reparsed, their temporary IDs no longer
// exist. In that case, restore the rule using its stable editor properties:
// file, selector, and parent.
//
// The original target is preserved so the selected preview element remains
// selected as well.
func (s *Service) restoreRuleSelection(
	target dto.SkinEditorTarget,
	oldActiveRule *dto.CSSRuleEditor,
	rules []parser.Rule,
) error {
	var selected *parser.Rule

	// First try the exact parser ID. This is the normal path for rules that
	// already existed on disk before the reload.
	if target.RuleID != "" {
		if rule, ok := editor.FindRuleByID(
			rules,
			target.RuleID,
		); ok {
			selected = rule
		}
	}

	// A newly-created rule receives a temporary ":new:" ID. After save and
	// reload that ID is gone, so find the corresponding parsed rule using the
	// stable properties that survived the save.
	if selected == nil {
		selected = findRuleForRestoration(
			rules,
			target,
			oldActiveRule,
		)
	}

	if selected == nil {
		return fmt.Errorf(
			"rule %q could not be restored",
			target.RuleID,
		)
	}

	active := editor.NewCSSRuleEditor(
		selected,
	)

	if active == nil {
		return fmt.Errorf(
			"rule %q could not be converted to editor state",
			selected.ID,
		)
	}

	// Preserve the selected element while updating the rule identity to the
	// newly parsed rule ID.
	restoredTarget := target

	restoredTarget.File = selected.File
	restoredTarget.RuleID = selected.ID
	restoredTarget.ParentID = selected.ParentID
	restoredTarget.Selector = selected.Selector
	restoredTarget.Mode = "existing"

	s.editor.SetTarget(
		restoredTarget,
	)

	s.editor.SetActiveRule(
		active,
	)

	logger.Infof(
		logModule,
		"restored rule selection: oldID=%q newID=%q file=%q selector=%q element=%q",
		target.RuleID,
		selected.ID,
		selected.File,
		selected.Selector,
		restoredTarget.ElementID,
	)

	return nil
}

// findRuleForRestoration finds the parsed rule corresponding to an editor
// selection whose original rule ID may no longer exist.
//
// Matching is intentionally conservative:
//
//   - file must match
//   - selector must match
//   - parent must match when the previous target had a parent
//
// The active rule is used as a fallback source for selector/file information
// when the target does not contain enough information.
func findRuleForRestoration(
	rules []parser.Rule,
	target dto.SkinEditorTarget,
	oldActiveRule *dto.CSSRuleEditor,
) *parser.Rule {
	file := target.File
	selector := strings.TrimSpace(target.Selector)
	parentID := target.ParentID

	if oldActiveRule != nil {
		if file == "" {
			file = oldActiveRule.File
		}

		if selector == "" {
			selector = strings.TrimSpace(
				oldActiveRule.Selector,
			)
		}

		if parentID == "" {
			parentID = oldActiveRule.ParentID
		}
	}

	if file == "" || selector == "" {
		return nil
	}

	var match *parser.Rule

	editor.WalkRules(
		rules,
		func(rule parser.Rule) bool {
			if rule.Selector == "" {
				return false
			}

			if rule.File != file {
				return false
			}

			if strings.TrimSpace(rule.Selector) != selector {
				return false
			}

			// ParentID is the strongest additional identity available for
			// rules nested inside the same container.
			if parentID != "" && rule.ParentID != parentID {
				return false
			}

			copy := rule
			match = &copy

			return true
		},
	)

	return match
}

// parseEditorRules parses all CSS files in the editor working copy.
//
// Each parsed file contributes its top-level rules to the single working-copy
// rule tree. Nested at-rules remain nested within their parser Rule values.
func parseEditorRules(
	files []dto.CSSFile,
) []parser.Rule {
	var rules []parser.Rule

	for _, file := range files {
		if filepath.Ext(file.Path) != ".css" {
			continue
		}

		rules = append(
			rules,
			parser.Parse(
				file.Path,
				file.Contents,
			)...,
		)
	}

	return rules
}

// logEditorRules logs the complete parsed rule tree for debugging.
func logEditorRules(
	rules []parser.Rule,
) {
	for i := range rules {
		logEditorRule(
			&rules[i],
			0,
			i,
		)
	}
}

// logEditorRule logs one rule and recursively logs its children.
func logEditorRule(
	rule *parser.Rule,
	depth int,
	index int,
) {
	indent := ""

	for i := 0; i < depth; i++ {
		indent += " "
	}

	logger.Infof(
		logModule,
		"%srule[%d] file=%q selector=%q layer=%q atRule=%q children=%d prelude=%t",
		indent,
		index,
		rule.File,
		rule.Selector,
		rule.Layer,
		rule.AtRule,
		len(rule.Children),
		rule.Prelude,
	)

	for i := range rule.Children {
		logEditorRule(
			&rule.Children[i],
			depth+2,
			i,
		)
	}
}
