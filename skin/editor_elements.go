package skin

import (
	"sort"
	"strings"

	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/skin/parser"
)

// SetPreviewElements replaces the set of preview elements supplied by the
// frontend.
func (s *Service) SetPreviewElements(
	elements []dto.SkinPreviewElement,
) error {
	s.editor.SetPreviewElements(
		elements,
	)

	return s.EmitSkinModel()
}

// GetSkinElements returns the combined list of CSS-defined elements and
// frontend preview elements.
//
// CSS selectors take precedence when both sources use the same element ID.
func (s *Service) GetSkinElements() []dto.SkinElement {
	cssElements := s.GetSkinSelectors()
	previewElements := s.editor.GetPreviewElements()

	seen := make(
		map[string]bool,
		len(cssElements)+len(previewElements),
	)

	elements := make(
		[]dto.SkinElement,
		0,
		len(cssElements)+len(previewElements),
	)

	for _, element := range cssElements {
		elements = append(
			elements,
			element,
		)

		seen[element.ID] = true
	}

	for _, element := range previewElements {
		if seen[element.ID] {
			continue
		}

		elements = append(
			elements,
			dto.SkinElement{
				ID:       element.ID,
				Label:    element.Label,
				Selector: element.Selector,
			},
		)
	}

	return elements
}

// GetSkinSelectors returns the unique CSS selectors defined by the active
// skin.
func (s *Service) GetSkinSelectors() []dto.SkinElement {
	_, rules, _, _, _, _ := s.editor.Snapshot()

	seen := make(
		map[string]bool,
	)

	elements := make(
		[]dto.SkinElement,
		0,
	)

	collectSelectors(
		rules,
		seen,
		&elements,
	)

	sort.Slice(
		elements,
		func(i, j int) bool {
			return elements[i].Selector <
				elements[j].Selector
		},
	)

	return elements
}

// collectSelectors recursively collects unique concrete CSS selectors from
// a parsed rule tree.
//
// Container rules such as @media, @supports, and @layer do not contribute
// elements themselves because they have no concrete Selector value. Their
// child rules are traversed recursively.
func collectSelectors(
	rules []parser.Rule,
	seen map[string]bool,
	elements *[]dto.SkinElement,
) {
	for _, rule := range rules {
		selector := strings.TrimSpace(
			rule.Selector,
		)

		if selector != "" &&
			!seen[selector] {
			seen[selector] = true

			*elements = append(
				*elements,
				dto.SkinElement{
					ID:       selector,
					Label:    selector,
					Selector: selector,
					File:     rule.File,
					Line:     rule.Line,
					Layer:    rule.Layer,
				},
			)
		}

		collectSelectors(
			rule.Children,
			seen,
			elements,
		)
	}
}
