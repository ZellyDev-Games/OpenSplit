package skin

import (
	"fmt"

	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/skin/editor"
	"github.com/zellydev-games/opensplit/skin/parser"
)

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

// findRule searches the complete rule tree for a rule belonging to the
// requested file and rule ID.
func findRule(
	rules []parser.Rule,
	file string,
	ruleID string,
) (*parser.Rule, bool) {
	var found *parser.Rule

	walkRules(
		rules,
		func(rule parser.Rule) bool {
			if rule.File != file ||
				rule.ID != ruleID {
				return false
			}

			copy := rule

			found = &copy

			return true
		},
	)

	return found, found != nil
}

// newCSSRuleEditor converts a parser rule into the DTO consumed by the
// frontend rule editor.
func newCSSRuleEditor(
	rule *parser.Rule,
) *dto.CSSRuleEditor {
	if rule == nil {
		return nil
	}

	return &dto.CSSRuleEditor{
		ID: rule.ID,

		File: rule.File,

		Selector: rule.Selector,

		Layer: rule.Layer,

		ParentID: rule.ParentID,

		Body: editor.FormatRuleDeclarations(
			rule.Declarations,
		),

		OriginalFile: rule.File,

		OriginalID: rule.ID,
	}
}

// parseDeclarations converts editor body text into structured parser
// declarations.
//
// The frontend editor works with declaration text only. Wrapping that
// text in a temporary selector lets the existing CSS parser perform the
// syntax handling.
func parseDeclarations(
	body string,
) []parser.Declaration {
	rules := parser.Parse(
		"editor",
		".temporary {\n"+body+"\n}",
	)

	if len(rules) == 0 {
		return nil
	}

	return rules[0].Declarations
}
