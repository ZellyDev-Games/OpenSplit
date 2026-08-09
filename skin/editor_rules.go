package skin

import (
	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/skin/editor"
	"github.com/zellydev-games/opensplit/skin/parser"
)

// convertDTOEditorRules converts parser rules into the DTO representation
// consumed by the frontend.
//
// Conversion is recursive so nested @media, @supports, @layer, and other
// rule containers retain their complete rule tree.
func convertDTOEditorRules(
	rules []parser.Rule,
) []dto.CSSRule {
	if rules == nil {
		return nil
	}

	out := make(
		[]dto.CSSRule,
		0,
		len(rules),
	)

	for _, rule := range rules {
		out = append(
			out,
			convertDTOEditorRule(rule),
		)
	}

	return out
}

// convertDTOEditorRule converts a single parser rule into its frontend DTO.
//
// Child rules are converted recursively so the complete CSS rule hierarchy
// remains available to the frontend.
func convertDTOEditorRule(
	rule parser.Rule,
) dto.CSSRule {
	return dto.CSSRule{
		ID: rule.ID,

		File: rule.File,

		Layer: rule.Layer,

		Selector: rule.Selector,

		Body: editor.FormatRuleDeclarations(
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

		ParentID: rule.ParentID,
	}
}

// findRuleByID searches the complete parsed rule tree for a rule ID.
//
// A copy of the matching rule is returned so callers cannot mutate the
// EditorState's internal rule tree.
func findRuleByID(
	rules []parser.Rule,
	ruleID string,
) (*parser.Rule, bool) {
	if ruleID == "" {
		return nil, false
	}

	var found *parser.Rule

	walkRules(
		rules,
		func(rule parser.Rule) bool {
			if rule.ID != ruleID {
				return false
			}

			copy := rule
			found = &copy

			return true
		},
	)

	return found, found != nil
}
