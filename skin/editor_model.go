package skin

import (
	"fmt"
	"path/filepath"
	"strings"

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

	files, err :=
		s.buildEditorFiles()

	if err != nil {
		return err
	}

	var rules []parser.Rule

	for _, file := range files {

		if filepath.Ext(file.Path) != ".css" {
			continue
		}

		parsed :=
			parser.Parse(
				file.Path,
				file.Contents,
			)

		rules =
			append(
				rules,
				parsed...,
			)
	}
	oldTarget :=
		s.editor.GetTarget()

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

	for i, rule := range rules {
		logger.Infof(
			logModule,
			"rule[%d] selectors=%v children=%d atRule=%q prelude=%t",
			i,
			rule.Selector,
			len(rule.Children),
			rule.AtRule,
			rule.Prelude,
		)
	}

	for i, rule := range rules {
		logger.Infof(logModule,
			"rule[%d] children=%d",
			i,
			len(rule.Children),
		)

		for j, child := range rule.Children {
			logger.Infof(logModule,
				" child[%d] selectors=%v atRule=%q children=%d",
				j,
				child.Selector,
				child.AtRule,
				len(child.Children),
			)
		}
	}

	switch {
	case oldTarget.RuleID != "":
		if err := s.restoreRuleSelection(oldTarget, rules); err != nil {
			s.editor.SetTarget(dto.SkinEditorTarget{
				Mode: "file",
				File: oldTarget.File,
			})
			s.editor.SetActiveRule(nil)
		}

	case oldTarget.File != "":
		s.editor.SetTarget(dto.SkinEditorTarget{
			Mode: "file",
			File: oldTarget.File,
		})
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

	logger.Infof(logModule, "editor elements: %d", len(s.GetSkinElements()))

	return SkinModel{

		Name: s.SelectedSkin(),

		Skins: s.GetAvailableSkins(),

		Directory: s.GetSkinPath(),

		StyleSheet: s.GetSkinAddress(),

		Files: files,

		Elements: s.GetSkinElements(),

		Rules: convertDTOEditorRules(
			rules,
		),

		Target: target,

		ActiveRule: active,

		Dirty: dirty,

		Revision: revision,
	}, nil
}

// EmitSkinModel sends the current cached editor state.
func (s *Service) EmitSkinModel() error {

	model, err :=
		s.GetSkinEditorModel()

	if err != nil {
		return err
	}

	select {

	case s.skinModelCh <- model:

	default:

	}

	return nil
}

func (s *Service) restoreRuleSelection(
	target dto.SkinEditorTarget,
	rules []parser.Rule,
) error {

	var found *parser.Rule

	walkRules(
		rules,
		func(rule parser.Rule) bool {

			if rule.ID == target.RuleID {

				copy := rule
				found = &copy

				return true
			}

			return false
		},
	)

	if found == nil {

		return fmt.Errorf(
			"rule %s no longer exists",
			target.RuleID,
		)
	}

	editor :=
		dto.CSSRuleEditor{

			ID: found.ID,

			File: found.File,

			Selector: "",

			Layer: found.Layer,

			Body: formatRuleDeclarations(
				found.Declarations,
			),

			OriginalFile: found.File,

			OriginalID: found.ID,
		}

	editor.Selector =
		found.Selector

	s.editor.SetTarget(
		target,
	)

	s.editor.SetActiveRule(
		&editor,
	)

	return nil
}

// buildEditorFiles reads editable files from the selected skin.
func (s *Service) buildEditorFiles() (
	[]dto.CSSFile,
	error,
) {

	files, err :=
		s.SkinFiles()

	if err != nil {
		return nil, err
	}

	out :=
		make(
			[]dto.CSSFile,
			0,
			len(files),
		)

	for _, file := range files {

		contents := ""

		text :=
			isTextFile(file)

		if text {

			contents, err =
				s.ReadFile(file)

			logger.Infof(
				logModule,
				"%s length=%d",
				file,
				len(contents),
			)

			if err != nil {
				return nil, err
			}
		}

		base :=
			strings.TrimSuffix(
				s.GetSkinAddress(),
				"/"+EntryPoint,
			)

		rel :=
			filepath.ToSlash(file)

		out =
			append(
				out,
				dto.CSSFile{

					Name: filepath.Base(rel),

					Path: rel,

					Contents: contents,

					OriginalContents: contents,

					Text: text,

					URL: base + "/" + rel,

					Type: fileType(rel),
				},
			)
	}

	return out, nil
}

func convertDTOEditorRules(
	rules []parser.Rule,
) []dto.CSSRule {

	out :=
		make(
			[]dto.CSSRule,
			0,
			len(rules),
		)

	for _, rule := range rules {

		out =
			append(
				out,
				dto.CSSRule{

					ID: rule.ID,

					File: rule.File,

					Layer: rule.Layer,

					Selector: rule.Selector,

					Body: formatRuleDeclarations(
						rule.Declarations,
					),

					AtRule: rule.AtRule,

					Block: rule.Block,

					Prelude: rule.Prelude,

					Children: convertDTOEditorRules(
						rule.Children,
					),

					Line: rule.Line,

					Order: rule.Order,
				},
			)
	}

	return out
}

func formatRuleDeclarations(
	declarations []parser.Declaration,
) string {

	var out strings.Builder

	for _, declaration := range declarations {

		for _, comment := range declaration.LeadingComment {

			out.WriteString(
				comment,
			)

			out.WriteString(
				"\n",
			)
		}

		if declaration.Name == "" {

			for _, raw := range declaration.Raw {

				out.WriteString(
					raw,
				)

				out.WriteString(
					"\n",
				)
			}

			continue
		}

		out.WriteString(
			declaration.Name,
		)

		out.WriteString(
			": ",
		)

		if len(declaration.Value) > 0 {

			out.WriteString(
				strings.Join(
					declaration.Value,
					"\n",
				),
			)

			if !strings.HasSuffix(
				strings.TrimSpace(
					declaration.Value[len(declaration.Value)-1],
				),
				";",
			) {
				out.WriteString(";")
			}
		}

		if declaration.InlineComment != "" {

			out.WriteString(
				" ",
			)

			out.WriteString(
				declaration.InlineComment,
			)
		}

		out.WriteString(
			"\n",
		)
	}

	return strings.TrimSpace(
		out.String(),
	)
}

func fileType(
	path string,
) string {

	ext :=
		strings.ToLower(
			filepath.Ext(path),
		)

	switch ext {

	case ".css":
		return "css"

	case ".png",
		".jpg",
		".jpeg",
		".gif",
		".webp":
		return "image"

	case ".svg":
		return "svg"

	case ".txt",
		".md",
		".html",
		".js",
		".json":
		return "text"
	}

	return "binary"
}

func isTextFile(
	path string,
) bool {

	switch strings.ToLower(filepath.Ext(path)) {

	case ".css",
		".txt",
		".md",
		".html",
		".js",
		".json":

		return true
	}

	return false
}

func walkRules(
	rules []parser.Rule,
	fn func(parser.Rule) bool,
) bool {

	for _, rule := range rules {

		if fn(rule) {
			return true
		}

		if walkRules(
			rule.Children,
			fn,
		) {
			return true
		}
	}

	return false
}
