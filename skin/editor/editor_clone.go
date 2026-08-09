package editor

import (
	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/skin/parser"
)

// CloneFiles returns an independent copy of the file slice.
func CloneFiles(
	files []dto.CSSFile,
) []dto.CSSFile {
	if files == nil {
		return nil
	}

	out := make(
		[]dto.CSSFile,
		len(files),
	)

	copy(
		out,
		files,
	)

	return out
}

// ClonePreviewElements returns an independent copy of preview elements.
func ClonePreviewElements(
	elements []dto.SkinPreviewElement,
) []dto.SkinPreviewElement {
	if elements == nil {
		return nil
	}

	out := make(
		[]dto.SkinPreviewElement,
		len(elements),
	)

	copy(
		out,
		elements,
	)

	return out
}

// CloneDeclarations returns a deep copy of declaration data.
func CloneDeclarations(
	declarations []parser.Declaration,
) []parser.Declaration {
	if declarations == nil {
		return nil
	}

	out := make(
		[]parser.Declaration,
		len(declarations),
	)

	for i, declaration := range declarations {
		out[i] = declaration

		out[i].Value = append(
			[]string{},
			declaration.Value...,
		)

		out[i].Raw = append(
			[]string{},
			declaration.Raw...,
		)

		out[i].LeadingComment = append(
			[]string{},
			declaration.LeadingComment...,
		)
	}

	return out
}

// CloneRules returns a deep copy of the complete rule tree.
func CloneRules(
	rules []parser.Rule,
) []parser.Rule {
	if rules == nil {
		return nil
	}

	out := make(
		[]parser.Rule,
		len(rules),
	)

	for i, rule := range rules {
		out[i] = rule

		out[i].Declarations = CloneDeclarations(
			rule.Declarations,
		)

		out[i].Children = CloneRules(
			rule.Children,
		)
	}

	return out
}
