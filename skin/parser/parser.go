package parser

import (
	"fmt"
	"os"
	"strings"
)

// ParseFile reads and parses a CSS file.
func ParseFile(
	filename string,
) ([]Rule, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return Parse(
		filename,
		string(data),
	), nil
}

// Parse parses CSS into the parser's rule tree.
//
// Rule IDs are assigned after the complete tree has been built so that
// callers can use stable IDs while editing the parsed working copy.
//
// Parent IDs are assigned at the same time because parent rule IDs cannot
// be known until the complete rule tree has been constructed.
func Parse(
	filename string,
	css string,
) []Rule {
	rules := parse(
		filename,
		css,
		1,
		"",
	)

	assignRuleIDs(
		rules,
		filename,
		"",
	)

	return rules
}

// assignRuleIDs assigns IDs, parent IDs, and sibling order recursively.
//
// IDs are based on the source filename, source line, and sibling index. The
// editor uses these IDs to identify rules across its working-copy operations.
//
// parentID is the ID of the rule containing the current rule. Root rules
// have an empty ParentID because they have no parent rule.
func assignRuleIDs(
	rules []Rule,
	filename string,
	parentID string,
) {
	for index := range rules {
		rule := &rules[index]

		rule.Order = index

		rule.ID = fmt.Sprintf(
			"%s:%d:%d",
			filename,
			rule.Line,
			index,
		)

		rule.ParentID = parentID

		assignRuleIDs(
			rule.Children,
			filename,
			rule.ID,
		)
	}
}

// parse parses one CSS scope.
//
// layer contains the inherited @layer name. Nested at-rules create child
// scopes while ordinary selector rules are emitted into the current scope.
func parse(
	filename string,
	css string,
	startLine int,
	layer string,
) []Rule {
	var rules []Rule

	index := 0
	line := startLine

	for index < len(css) {
		index, line = skipWhitespaceAndComments(
			css,
			index,
			line,
		)

		if index >= len(css) {
			break
		}

		if css[index] == '@' {
			if rule, next, nextLine, ok := parseAtRuleStatement(
				filename,
				css,
				index,
				line,
				layer,
			); ok {
				rules = append(
					rules,
					rule,
				)

				index = next
				line = nextLine

				continue
			}
		}

		brace := findNextOpeningBrace(
			css,
			index,
		)

		if brace < 0 {
			break
		}

		closeBrace := findMatchingBrace(
			css,
			brace,
		)

		if closeBrace < 0 {
			break
		}

		header := strings.TrimSpace(
			css[index:brace],
		)

		if header == "" {
			index = closeBrace + 1

			continue
		}

		headerLine := line

		line += strings.Count(
			css[index:closeBrace+1],
			"\n",
		)

		body := css[brace+1 : closeBrace]

		rules = append(
			rules,
			parseBlock(
				filename,
				header,
				body,
				headerLine,
				layer,
			)...,
		)

		index = closeBrace + 1
	}

	return rules
}

// parseAtRuleStatement attempts to parse a non-block at-rule.
//
// It returns ok=false when the at-rule actually begins a block, allowing the
// normal block parser to handle it.
func parseAtRuleStatement(
	filename string,
	css string,
	start int,
	line int,
	layer string,
) (Rule, int, int, bool) {
	end, ok := findStatementEnd(
		css,
		start,
	)

	if !ok {
		return Rule{}, start, line, false
	}

	text := strings.TrimSpace(
		css[start : end+1],
	)

	rule := Rule{
		File: filename,

		Layer: layer,

		AtRule: text,

		Block: false,

		Prelude: isPreludeAtRule(text),

		Line: line,
	}

	nextLine := line + strings.Count(
		text,
		"\n",
	)

	return rule, end + 1, nextLine, true
}

// parseBlock parses one complete CSS block.
//
// A block beginning with @ is parsed as an at-rule. Otherwise it is parsed
// as one or more selector rules.
func parseBlock(
	filename string,
	header string,
	body string,
	headerLine int,
	layer string,
) []Rule {
	if strings.HasPrefix(
		header,
		"@",
	) {
		return []Rule{
			parseAtRule(
				filename,
				header,
				body,
				headerLine,
				layer,
			),
		}
	}

	return parseSelectorRules(
		filename,
		header,
		body,
		headerLine,
		layer,
	)
}

// parseAtRule constructs a block at-rule and recursively parses its contents
// when the at-rule contains nested rules.
func parseAtRule(
	filename string,
	header string,
	body string,
	headerLine int,
	layer string,
) Rule {
	rule := Rule{
		File: filename,

		Layer: layer,

		AtRule: strings.TrimSpace(
			header,
		),

		Block: true,

		Line: headerLine,
	}

	switch {
	case isLayerAtRule(header):
		rule.Children = parseLayerBody(
			filename,
			header,
			body,
			headerLine,
		)

	case isDeclarationAtRule(header):
		rule.Declarations = ParseDeclarations(
			body,
		)

	default:
		rule.Children = parse(
			filename,
			body,
			headerLine+1,
			layer,
		)
	}

	return rule
}

// parseLayerBody parses the contents of an @layer block.
//
// Rules inside a named layer inherit that layer name rather than the parent
// layer's name.
func parseLayerBody(
	filename string,
	header string,
	body string,
	headerLine int,
) []Rule {
	layer := layerNameFromAtRule(
		header,
	)

	return parse(
		filename,
		body,
		headerLine+1,
		layer,
	)
}

// parseSelectorRules parses a selector list into individual Rule values.
//
// A selector list such as:
//
//	.foo, .bar
//
// becomes two independent rules while sharing the same declaration set.
func parseSelectorRules(
	filename string,
	header string,
	body string,
	headerLine int,
	layer string,
) []Rule {
	declarations := ParseDeclarations(
		strings.TrimSpace(body),
	)

	selectors := splitSelectors(
		header,
	)

	rules := make(
		[]Rule,
		0,
		len(selectors),
	)

	for _, selector := range selectors {
		selector = normalizeSelector(
			selector,
		)

		if selector == "" {
			continue
		}

		rules = append(
			rules,
			Rule{
				File: filename,

				Layer: layer,

				Selector: selector,

				Declarations: cloneDeclarations(
					declarations,
				),

				Line: headerLine,
			},
		)
	}

	return rules
}

// skipWhitespaceAndComments advances past whitespace and block comments.
//
// Comments are intentionally skipped at the rule-tree level. Declaration
// comments are preserved separately by ParseDeclarations.
func skipWhitespaceAndComments(
	css string,
	index int,
	line int,
) (int, int) {
	for index < len(css) {
		index, line = skipWhitespace(
			css,
			index,
			line,
		)

		if index >= len(css) {
			return index, line
		}

		if !hasPrefixAt(
			css,
			index,
			"/*",
		) {
			break
		}

		next, nextLine, ok := skipBlockComment(
			css,
			index,
			line,
		)

		if !ok {
			return len(css), nextLine
		}

		index = next
		line = nextLine
	}

	return index, line
}

// skipWhitespace advances past CSS whitespace and updates the source line.
func skipWhitespace(
	css string,
	index int,
	line int,
) (int, int) {
	for index < len(css) {
		switch css[index] {
		case ' ', '\t', '\r':
			index++

		case '\n':
			index++
			line++

		default:
			return index, line
		}
	}

	return index, line
}

// skipBlockComment skips one CSS block comment and updates the source line.
func skipBlockComment(
	css string,
	index int,
	line int,
) (int, int, bool) {
	end := strings.Index(
		css[index+2:],
		"*/",
	)

	if end < 0 {
		return len(css), line, false
	}

	end += index + 4

	line += strings.Count(
		css[index:end],
		"\n",
	)

	return end, line, true
}

// findStatementEnd finds the semicolon terminating a non-block at-rule.
//
// Strings and comments are respected so that semicolons inside quoted
// strings do not terminate the statement.
func findStatementEnd(
	css string,
	start int,
) (int, bool) {
	inString := false
	var quote byte

	inComment := false

	for index := start; index < len(css); index++ {
		char := css[index]

		if inComment {
			if isCommentEnd(
				css,
				index,
			) {
				inComment = false
				index++
			}

			continue
		}

		if inString {
			if char == '\\' &&
				index+1 < len(css) {
				index++
				continue
			}

			if char == quote {
				inString = false
			}

			continue
		}

		if isCommentStart(
			css,
			index,
		) {
			inComment = true
			index++

			continue
		}

		if char == '"' ||
			char == '\'' {
			inString = true
			quote = char

			continue
		}

		switch char {
		case ';':
			return index, true

		case '{':
			return -1, false
		}
	}

	return -1, false
}

// findNextOpeningBrace finds the next block opening brace while ignoring
// braces inside strings and comments.
func findNextOpeningBrace(
	css string,
	start int,
) int {
	inString := false
	var quote byte

	inComment := false

	for index := start; index < len(css); index++ {
		char := css[index]

		if inComment {
			if isCommentEnd(
				css,
				index,
			) {
				inComment = false
				index++
			}

			continue
		}

		if inString {
			if char == '\\' &&
				index+1 < len(css) {
				index++
				continue
			}

			if char == quote {
				inString = false
			}

			continue
		}

		if isCommentStart(
			css,
			index,
		) {
			inComment = true
			index++

			continue
		}

		if char == '"' ||
			char == '\'' {
			inString = true
			quote = char

			continue
		}

		if char == '{' {
			return index
		}
	}

	return -1
}

// findMatchingBrace finds the closing brace corresponding to the opening
// brace at start.
//
// Braces inside strings and comments do not affect nesting.
func findMatchingBrace(
	css string,
	start int,
) int {
	depth := 1

	inString := false
	var quote byte

	inComment := false

	for index := start + 1; index < len(css); index++ {
		char := css[index]

		if inComment {
			if isCommentEnd(
				css,
				index,
			) {
				inComment = false
				index++
			}

			continue
		}

		if inString {
			if char == '\\' &&
				index+1 < len(css) {
				index++
				continue
			}

			if char == quote {
				inString = false
			}

			continue
		}

		if isCommentStart(
			css,
			index,
		) {
			inComment = true
			index++

			continue
		}

		if char == '"' ||
			char == '\'' {
			inString = true
			quote = char

			continue
		}

		switch char {
		case '{':
			depth++

		case '}':
			depth--

			if depth == 0 {
				return index
			}
		}
	}

	return -1
}

// isCommentStart reports whether a block comment starts at index.
func isCommentStart(
	text string,
	index int,
) bool {
	return hasPrefixAt(
		text,
		index,
		"/*",
	)
}

// isCommentEnd reports whether a block comment ends at index.
func isCommentEnd(
	text string,
	index int,
) bool {
	return hasPrefixAt(
		text,
		index,
		"*/",
	)
}

// isPreludeAtRule identifies at-rules that are represented as standalone
// statements and may participate in CSS preprocessing.
func isPreludeAtRule(
	text string,
) bool {
	return strings.HasPrefix(
		strings.ToLower(
			strings.TrimSpace(text),
		),
		"@import",
	)
}

// isLayerAtRule identifies block @layer rules.
func isLayerAtRule(
	header string,
) bool {
	return strings.HasPrefix(
		strings.ToLower(
			strings.TrimSpace(header),
		),
		"@layer",
	)
}

// layerNameFromAtRule returns the layer name from a block @layer header.
//
// Examples:
//
//	@layer base
//	    => base
//
//	@layer components
//	    => components
//
//	@LAYER utilities
//	    => utilities
//
// Anonymous @layer blocks return an empty layer name.
func layerNameFromAtRule(
	header string,
) string {
	header = strings.TrimSpace(
		header,
	)

	if len(header) < len("@layer") {
		return ""
	}

	return strings.TrimSpace(
		header[len("@layer"):],
	)
}

// isDeclarationAtRule identifies block at-rules whose contents are
// declarations rather than nested rules.
func isDeclarationAtRule(
	header string,
) bool {
	header = strings.ToLower(
		strings.TrimSpace(header),
	)

	switch {
	case strings.HasPrefix(
		header,
		"@font-face",
	):
		return true

	case strings.HasPrefix(
		header,
		"@page",
	):
		return true

	default:
		return false
	}
}

// normalizeSelector normalizes whitespace in a selector while preserving
// the selector's semantic punctuation.
func normalizeSelector(
	selector string,
) string {
	return strings.Join(
		strings.Fields(selector),
		" ",
	)
}

// cloneDeclarations creates a deep-enough copy of declarations for rule
// splitting. Slice fields are copied so sibling rules do not share mutable
// declaration slices.
func cloneDeclarations(
	in []Declaration,
) []Declaration {
	out := make(
		[]Declaration,
		len(in),
	)

	for index, declaration := range in {
		out[index] = declaration

		out[index].Value = append(
			[]string{},
			declaration.Value...,
		)

		out[index].LeadingComment = append(
			[]string{},
			declaration.LeadingComment...,
		)

		out[index].Raw = append(
			[]string{},
			declaration.Raw...,
		)
	}

	return out
}

// splitSelectors splits a selector list while respecting (), [] and {}.
//
// Commas inside functional pseudo selectors or attribute selectors are not
// treated as selector-list separators.
func splitSelectors(
	header string,
) []string {
	var selectors []string
	var current strings.Builder

	parens := 0
	brackets := 0
	braces := 0

	for _, char := range header {
		switch char {
		case '(':
			parens++

		case ')':
			if parens > 0 {
				parens--
			}

		case '[':
			brackets++

		case ']':
			if brackets > 0 {
				brackets--
			}

		case '{':
			braces++

		case '}':
			if braces > 0 {
				braces--
			}

		case ',':
			if parens == 0 &&
				brackets == 0 &&
				braces == 0 {
				appendSelector(
					&selectors,
					current.String(),
				)

				current.Reset()

				continue
			}
		}

		current.WriteRune(char)
	}

	appendSelector(
		&selectors,
		current.String(),
	)

	return selectors
}

// appendSelector appends a non-empty selector to a selector list.
func appendSelector(
	out *[]string,
	selector string,
) {
	selector = strings.TrimSpace(
		selector,
	)

	if selector == "" {
		return
	}

	*out = append(
		*out,
		selector,
	)
}

// hasPrefixAt reports whether text contains prefix at the supplied index.
func hasPrefixAt(
	text string,
	index int,
	prefix string,
) bool {
	return index >= 0 &&
		index+len(prefix) <= len(text) &&
		text[index:index+len(prefix)] == prefix
}
