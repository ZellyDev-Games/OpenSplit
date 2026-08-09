package editor

import (
	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/skin/parser"
)

// ConvertRules converts parser rules into the DTO representation consumed by
// the frontend.
func ConvertRules(
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
			ConvertRule(rule),
		)
	}

	return out
}

// ConvertRule converts a single parser rule into its frontend DTO.
func ConvertRule(
	rule parser.Rule,
) dto.CSSRule {
	return dto.CSSRule{
		ID: rule.ID,

		File: rule.File,

		Layer: rule.Layer,

		Selector: rule.Selector,

		Body: FormatRuleDeclarations(
			rule.Declarations,
		),

		AtRule: rule.AtRule,

		Block: rule.Block,

		Prelude: rule.Prelude,

		Children: ConvertRules(
			rule.Children,
		),

		Line: rule.Line,

		Order: rule.Order,

		ParentID: rule.ParentID,
	}
}
