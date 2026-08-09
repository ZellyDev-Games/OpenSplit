package skin

import (
	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/skin/parser"
)

// cloneFiles returns an independent copy of the file slice.
//
// CSSFile currently contains value fields, so copying each element is
// sufficient.
func cloneFiles(
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

// clonePreviewElements returns an independent copy of preview elements.
func clonePreviewElements(
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

// cloneDeclarations returns a deep copy of declaration data.
func cloneDeclarations(
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

// cloneRules returns a deep copy of the complete rule tree.
func cloneRules(
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

		out[i].Declarations = cloneDeclarations(
			rule.Declarations,
		)

		out[i].Children = cloneRules(
			rule.Children,
		)
	}

	return out
}
