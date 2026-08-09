package skin

import (
	"fmt"
	"path/filepath"

	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/logger"
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

	oldTarget := s.editor.GetTarget()

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
			rules,
		); err != nil {
			s.editor.SetTarget(
				dto.SkinEditorTarget{
					Mode: "file",
					File: oldTarget.File,
				},
			)

			s.editor.SetActiveRule(nil)
		}

	case oldTarget.File != "":
		s.editor.SetTarget(
			dto.SkinEditorTarget{
				Mode: "file",
				File: oldTarget.File,
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

		Rules: convertDTOEditorRules(
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
// The rule ID is used as the stable identity. If the rule no longer exists,
// the caller can fall back to file selection.
func (s *Service) restoreRuleSelection(
	target dto.SkinEditorTarget,
	rules []parser.Rule,
) error {
	selected, ok := findRuleByID(
		rules,
		target.RuleID,
	)

	if !ok {
		return fmt.Errorf(
			"rule %s no longer exists",
			target.RuleID,
		)
	}

	editor := newCSSRuleEditor(
		selected,
	)

	if editor == nil {
		return fmt.Errorf(
			"rule %s could not be converted to editor state",
			target.RuleID,
		)
	}

	s.editor.SetTarget(target)
	s.editor.SetActiveRule(editor)

	return nil
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
