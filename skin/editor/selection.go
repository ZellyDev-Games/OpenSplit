package editor

import (
	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/skin/parser"
)

// FindRule searches the complete rule tree for a rule belonging to the
// requested file and rule ID.
func FindRule(
	rules []parser.Rule,
	file string,
	ruleID string,
) (*parser.Rule, bool) {
	var found *parser.Rule

	WalkRules(
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

// FindRuleByID searches the complete parsed rule tree for a rule ID.
func FindRuleByID(
	rules []parser.Rule,
	ruleID string,
) (*parser.Rule, bool) {
	if ruleID == "" {
		return nil, false
	}

	var found *parser.Rule

	WalkRules(
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

// NewCSSRuleEditor converts a parser rule into the DTO consumed by the
// frontend rule editor.
func NewCSSRuleEditor(
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

		Body: FormatRuleDeclarations(
			rule.Declarations,
		),

		OriginalFile: rule.File,

		OriginalID: rule.ID,
	}
}

// ParseDeclarations converts editor body text into structured parser
// declarations.
func ParseDeclarations(
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
