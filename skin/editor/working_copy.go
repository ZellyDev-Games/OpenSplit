package editor

import (
	"github.com/zellydev-games/opensplit/skin/parser"
)

// InsertRuleRecursive inserts a rule into the requested parent.
func InsertRuleRecursive(
	rules []parser.Rule,
	parentID string,
	rule parser.Rule,
) ([]parser.Rule, bool) {
	if parentID == "" {
		return append(
			rules,
			rule,
		), true
	}

	for i := range rules {
		if rules[i].ID == parentID {
			rules[i].Children = append(
				rules[i].Children,
				rule,
			)

			return rules, true
		}

		updated, inserted := InsertRuleRecursive(
			rules[i].Children,
			parentID,
			rule,
		)

		if inserted {
			rules[i].Children = updated

			return rules, true
		}
	}

	return rules, false
}
