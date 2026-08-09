package skin

import "github.com/zellydev-games/opensplit/skin/parser"

// walkRules recursively visits every rule in a parsed CSS rule tree.
//
// Returning true from fn stops traversal immediately.
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
