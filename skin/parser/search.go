package parser

import (
	"slices"
	"strings"
)

// FindRules returns all parsed CSS rules matching a selector.
//
// Searches recursively through:
//   - @layer
//   - @media
//   - @supports
//   - @font-face
//   - other nested at-rules
func FindRules(
	filename string,
	css string,
	selector string,
) []Rule {
	rules :=
		Parse(
			filename,
			css,
		)

	var matches []Rule

	walkRules(
		rules,
		func(rule Rule) {
			if SelectorMatch(
				rule.Selector,
				selector,
			) {

				matches =
					append(
						matches,
						rule,
					)
			}
		},
	)

	return matches
}

// walkRules recursively visits every rule in the CSS tree.
func walkRules(
	rules []Rule,
	fn func(Rule),
) {

	for _, rule := range rules {

		fn(rule)

		walkRules(
			rule.Children,
			fn,
		)
	}
}

// SelectorMatch determines whether a CSS selector applies
// to the selected editor element.
//
// Matching is token based.
//
// Examples:
//
// .segment
//
// matches:
//
// .segment.active
// .segment.gold
// div.segment[data-state="active"]
func SelectorMatch(
	ruleSelector string,
	selected string,
) bool {

	ruleTokens :=
		selectorTokens(ruleSelector)

	selectedTokens :=
		selectorTokens(selected)

	for _, selectedToken := range selectedTokens {

		if !slices.Contains(
			ruleTokens,
			selectedToken,
		) {

			return false
		}
	}

	return true
}

// selectorTokens extracts comparable selector tokens.
//
// Examples:
//
// .segment.active
// => [.segment .active]
//
// div.segment
// => [div .segment]
//
// .segment[data-state="active"]
// => [.segment [data-state="active"]]
func selectorTokens(
	selector string,
) []string {

	replacer :=
		strings.NewReplacer(
			">",
			" ",

			"+",
			" ",

			"~",
			" ",

			",",
			" ",
		)

	selector =
		replacer.Replace(
			selector,
		)

	var tokens []string

	for _, field := range strings.Fields(selector) {

		// Ignore pseudo selectors.
		//
		// .segment:hover
		// becomes:
		//
		// .segment
		if i :=
			strings.IndexRune(
				field,
				':',
			); i >= 0 {

			field =
				field[:i]
		}

		if field == "" {
			continue
		}

		var current strings.Builder

		for _, r := range field {

			switch r {

			case '.', '#', '[':

				if current.Len() > 0 {

					tokens =
						append(
							tokens,
							current.String(),
						)

					current.Reset()
				}

				current.WriteRune(
					r,
				)

			default:

				current.WriteRune(
					r,
				)
			}
		}

		if current.Len() > 0 {

			tokens =
				append(
					tokens,
					current.String(),
				)
		}
	}

	return tokens
}
