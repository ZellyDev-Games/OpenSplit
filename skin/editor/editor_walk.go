package editor

import "github.com/zellydev-games/opensplit/skin/parser"

// walkRules recursively visits every rule in a parsed CSS rule tree.
//
// Returning true from fn stops traversal immediately.
func WalkRules(
	rules []parser.Rule,
	fn func(parser.Rule) bool,
) bool {
	for _, rule := range rules {
		if fn(rule) {
			return true
		}

		if WalkRules(
			rule.Children,
			fn,
		) {
			return true
		}
	}

	return false
}
