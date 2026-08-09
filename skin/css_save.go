package skin

import (
	"os"
	"path/filepath"

	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/skin/parser"
)

// SaveWorkingCopy persists the current editor working copy to disk.
//
// Saving is intentionally a separate operation from editor mutation.
// Editor changes only modify EditorState. This method is the single
// working-copy -> disk boundary.
//
// Saving does not reload the editor from disk. The existing working copy,
// selection, active rule, and editor UI state remain intact after the save.
func (s *Service) SaveWorkingCopy() error {
	_, rules, dirty, _, _, revision := s.editor.Snapshot()

	logger.Infof(
		logModule,
		"SaveWorkingCopy: rules=%d dirty=%t revision=%d",
		len(rules),
		dirty,
		revision,
	)

	if err := s.saveEditor(); err != nil {
		return err
	}

	if err := s.EmitSkinModel(); err != nil {
		return err
	}

	s.notifySkinUpdated()

	return nil
}

// saveEditor writes the current editor working copy to the selected skin.
//
// Text files are written using their current working-copy contents.
// CSS files are regenerated from the parser rule tree so mutations made
// through the CSS editor are persisted consistently.
func (s *Service) saveEditor() error {
	files, rules, _, _, _, _ := s.editor.Snapshot()

	root := s.GetSkinPath()

	rulesByFile := make(
		map[string][]parser.Rule,
	)

	for _, rule := range rules {
		if rule.File == "" {
			continue
		}

		rulesByFile[rule.File] = append(
			rulesByFile[rule.File],
			rule,
		)
	}

	for _, file := range files {
		if !file.Text {
			continue
		}

		contents := file.Contents

		if filepath.Ext(file.Path) == ".css" {
			fileRules := rulesByFile[file.Path]

			logger.Infof(
				logModule,
				"saveEditor: formatting CSS file=%q topLevelRules=%d",
				file.Path,
				len(fileRules),
			)

			for i, rule := range fileRules {
				logger.Infof(
					logModule,
					"saveEditor: rule[%d] id=%q selector=%q layer=%q atRule=%q children=%d parent=%q",
					i,
					rule.ID,
					rule.Selector,
					rule.Layer,
					rule.AtRule,
					len(rule.Children),
					rule.ParentID,
				)

				logRuleTree(
					rule,
					"  ",
				)
			}

			contents = parser.FormatCSS(
				fileRules,
			)

			logger.Infof(
				logModule,
				"saveEditor: formatted %q length=%d contents:\n%s",
				file.Path,
				len(contents),
				contents,
			)
		}

		filename := filepath.Join(
			root,
			file.Path,
		)

		if err := os.MkdirAll(
			filepath.Dir(filename),
			0o755,
		); err != nil {
			return err
		}

		logger.Infof(
			logModule,
			"saveEditor: writing %q length=%d",
			filename,
			len(contents),
		)

		if err := os.WriteFile(
			filename,
			[]byte(contents),
			0o644,
		); err != nil {
			return err
		}
	}

	s.editor.ClearDirty()

	return nil
}

// logRuleTree logs a rule and all descendants.
//
// This is intentionally kept local to the save path while debugging the
// working-copy -> CSS writer boundary.
func logRuleTree(
	rule parser.Rule,
	indent string,
) {
	logger.Infof(
		logModule,
		"%srule id=%q selector=%q layer=%q atRule=%q children=%d parent=%q",
		indent,
		rule.ID,
		rule.Selector,
		rule.Layer,
		rule.AtRule,
		len(rule.Children),
		rule.ParentID,
	)

	for _, child := range rule.Children {
		logRuleTree(
			child,
			indent+"  ",
		)
	}
}
