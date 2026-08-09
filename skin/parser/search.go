package parser

import (
	"slices"
	"strings"
)

// FindRules returns all parsed CSS rules matching a selector.
//
// Searches recursively through the complete CSS rule tree, including rules
// nested inside at-rules such as:
//
//	@layer
//	@media
//	@supports
//	@container
//	@keyframes
//
// The returned rules preserve their original parser order.
func FindRules(
	filename string,
	css string,
	selector string,
) []Rule {
	rules := Parse(
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
				matches = append(
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
	for i := range rules {
		fn(rules[i])

		walkRules(
			rules[i].Children,
			fn,
		)
	}
}

// SelectorMatch determines whether a CSS rule selector applies to the
// selected editor element.
//
// The editor does not attempt to implement the full CSS selector grammar.
// Instead, matching is intentionally limited to the selector tokens that
// identify an element:
//
//   - element names
//   - class selectors
//   - ID selectors
//   - attribute selectors
//
// Pseudo selectors are ignored because they describe element state rather
// than the identity of the preview element.
//
// A selector is considered a match when its final compound selector contains
// all of the identity tokens from the selected selector.
//
// Examples:
//
//	.segment
//	    matches .segment
//
//	.segment.active
//	    matches .segment.active
//
//	div.segment
//	    matches div.segment
//
//	.parent .segment
//	    matches .segment
//
//	.segment .time
//	    does not match .segment
func SelectorMatch(
	ruleSelector string,
	selected string,
) bool {
	ruleTokens := selectorTokens(
		lastSelectorCompound(
			ruleSelector,
		),
	)

	selectedTokens := selectorTokens(
		lastSelectorCompound(
			selected,
		),
	)

	if len(selectedTokens) == 0 {
		return false
	}

	return containsAllTokens(
		ruleTokens,
		selectedTokens,
	)
}

// containsAllTokens reports whether every selected token exists in the rule
// token set.
func containsAllTokens(
	ruleTokens []string,
	selectedTokens []string,
) bool {
	for _, token := range selectedTokens {
		if !slices.Contains(
			ruleTokens,
			token,
		) {
			return false
		}
	}

	return true
}

// lastSelectorCompound returns the final compound selector from a selector.
//
// Combinators separate compound selectors:
//
//	.parent .segment
//	    => .segment
//
//	.parent > .segment.active
//	    => .segment.active
//
//	.parent + .segment
//	    => .segment
//
// Attribute values and functional expressions are preserved because their
// internal whitespace and combinators do not represent selector structure.
//
// This is intentionally a limited parser for the skin editor rather than a
// complete CSS selector implementation.
func lastSelectorCompound(
	selector string,
) string {
	var current strings.Builder

	parens := 0
	brackets := 0

	for _, char := range selector {
		switch char {
		case '(':
			parens++
			current.WriteRune(char)

		case ')':
			if parens > 0 {
				parens--
			}

			current.WriteRune(char)

		case '[':
			brackets++
			current.WriteRune(char)

		case ']':
			if brackets > 0 {
				brackets--
			}

			current.WriteRune(char)

		case '>', '+', '~':
			if isSelectorSeparator(
				parens,
				brackets,
			) {
				current.Reset()
				continue
			}

			current.WriteRune(char)

		case ' ', '\t', '\r', '\n':
			if isSelectorSeparator(
				parens,
				brackets,
			) {
				current.Reset()
				continue
			}

			current.WriteRune(char)

		default:
			current.WriteRune(char)
		}
	}

	return strings.TrimSpace(
		current.String(),
	)
}

// isSelectorSeparator reports whether selector whitespace or a combinator is
// currently outside a protected selector expression.
func isSelectorSeparator(
	parens int,
	brackets int,
) bool {
	return parens == 0 &&
		brackets == 0
}

// selectorTokens extracts comparable selector tokens from a selector.
//
// Element, class, ID, and attribute tokens are retained:
//
//	div.segment#active[data-state="active"]
//	    => [div .segment #active [data-state="active"]]
//
// Pseudo selectors are discarded:
//
//	.segment:hover
//	    => [.segment]
//
// This function is intentionally not a complete CSS selector parser. It
// provides the limited identity matching required by the skin editor.
func selectorTokens(
	selector string,
) []string {
	selector = normalizeSelectorCombinators(
		selector,
	)

	var tokens []string

	for _, field := range strings.Fields(
		selector,
	) {
		field = stripPseudoSelector(
			field,
		)

		if field == "" {
			continue
		}

		tokens = append(
			tokens,
			splitSelectorField(field)...,
		)
	}

	return tokens
}

// normalizeSelectorCombinators replaces selector combinators with whitespace
// so they can be tokenized consistently.
func normalizeSelectorCombinators(
	selector string,
) string {
	return strings.NewReplacer(
		">", " ",
		"+", " ",
		"~", " ",
	).Replace(selector)
}

// stripPseudoSelector removes the pseudo-selector portion of a selector
// token.
//
// For example:
//
//	.segment:hover
//	    => .segment
//
//	.segment::before
//	    => .segment
//
// Attribute selectors are unaffected because ':' inside an attribute value
// is not treated as a pseudo selector here.
func stripPseudoSelector(
	field string,
) string {
	bracketDepth := 0
	inString := false

	var quote rune

	for index, char := range field {
		if inString {
			if char == '\\' {
				continue
			}

			if char == quote {
				inString = false
			}

			continue
		}

		switch char {
		case '"', '\'':
			if bracketDepth > 0 {
				inString = true
				quote = char
			}

		case '[':
			bracketDepth++

		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}

		case ':':
			if bracketDepth == 0 {
				return field[:index]
			}
		}
	}

	return field
}

// splitSelectorField separates the selector components that identify an
// element.
//
// For example:
//
//	div.segment#active[data-state="active"]
//
// becomes:
//
//	[div .segment #active [data-state="active"]]
//
// Attribute selectors are kept as a single token, including any '.' or '#'
// characters inside the attribute expression.
func splitSelectorField(
	field string,
) []string {
	var tokens []string
	var current strings.Builder

	bracketDepth := 0

	for _, char := range field {
		switch char {
		case '[':
			if bracketDepth == 0 {
				appendSelectorToken(
					&tokens,
					&current,
				)
			}

			bracketDepth++
			current.WriteRune(char)

		case ']':
			current.WriteRune(char)

			if bracketDepth > 0 {
				bracketDepth--

				if bracketDepth == 0 {
					appendSelectorToken(
						&tokens,
						&current,
					)
				}
			}

		case '.', '#':
			if bracketDepth == 0 {
				appendSelectorToken(
					&tokens,
					&current,
				)

				current.WriteRune(char)

				continue
			}

			current.WriteRune(char)

		default:
			current.WriteRune(char)
		}
	}

	appendSelectorToken(
		&tokens,
		&current,
	)

	return tokens
}

// appendSelectorToken appends the current selector token when it contains
// content.
func appendSelectorToken(
	tokens *[]string,
	current *strings.Builder,
) {
	if current.Len() == 0 {
		return
	}

	*tokens = append(
		*tokens,
		current.String(),
	)

	current.Reset()
}
