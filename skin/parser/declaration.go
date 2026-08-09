package parser

import "strings"

type parseState struct {
	parens   int
	brackets int

	inString bool
	quote    byte

	inComment bool
}

// ParseDeclarations parses the contents of a CSS declaration block.
//
// The parser preserves:
//
//   - declaration names and values
//   - multiline values
//   - leading comments
//   - inline comments
//   - raw/unsupported constructs
//
// Declaration boundaries are determined while respecting strings, comments,
// parentheses, and attribute brackets.
func ParseDeclarations(
	body string,
) []Declaration {
	body = normalizeLineEndings(body)

	var declarations []Declaration
	var pendingComments []string

	index := 0

	for index < len(body) {
		index = skipDeclarationWhitespace(
			body,
			index,
		)

		if index >= len(body) {
			break
		}

		if isDeclarationCommentStart(
			body,
			index,
		) {
			comment, next, ok := readComment(
				body,
				index,
			)

			if !ok {
				pendingComments = append(
					pendingComments,
					strings.TrimSpace(
						body[index:],
					),
				)

				break
			}

			pendingComments = append(
				pendingComments,
				comment,
			)

			index = next
			continue
		}

		colon := findDeclarationColon(
			body,
			index,
		)

		if colon < 0 {
			appendRawDeclaration(
				&declarations,
				pendingComments,
				body[index:],
			)

			break
		}

		name := strings.TrimSpace(
			body[index:colon],
		)

		if name == "" {
			index = colon + 1
			continue
		}

		valueStart := colon + 1

		valueEnd := findDeclarationEnd(
			body,
			valueStart,
		)

		value, next := declarationValue(
			body,
			valueStart,
			valueEnd,
		)

		inlineComment := extractInlineComment(
			value,
		)

		value = removeInlineComment(
			value,
		)

		declarations = append(
			declarations,
			Declaration{
				Name: name,

				Value: splitValueLines(
					value,
				),

				InlineComment: inlineComment,

				LeadingComment: pendingComments,
			},
		)

		pendingComments = nil
		index = next
	}

	return declarations
}

// normalizeLineEndings converts all supported line endings to LF.
func normalizeLineEndings(
	text string,
) string {
	text = strings.ReplaceAll(
		text,
		"\r\n",
		"\n",
	)

	return strings.ReplaceAll(
		text,
		"\r",
		"\n",
	)
}

// skipDeclarationWhitespace advances over whitespace between declarations.
func skipDeclarationWhitespace(
	body string,
	start int,
) int {
	for start < len(body) {
		switch body[start] {
		case ' ', '\t', '\r', '\n':
			start++

		default:
			return start
		}
	}

	return start
}

// isDeclarationCommentStart reports whether a block comment begins at index.
//
// This helper is intentionally declaration-parser-specific because parser.go
// owns the generic CSS comment scanner used by the rule parser.
func isDeclarationCommentStart(
	text string,
	index int,
) bool {
	return index >= 0 &&
		index+1 < len(text) &&
		text[index] == '/' &&
		text[index+1] == '*'
}

// readComment reads one CSS block comment.
//
// The returned next index points immediately after the closing */.
func readComment(
	text string,
	start int,
) (string, int, bool) {
	if !isDeclarationCommentStart(
		text,
		start,
	) {
		return "", start, false
	}

	end := strings.Index(
		text[start+2:],
		"*/",
	)

	if end < 0 {
		return "", start, false
	}

	end += start + 4

	return strings.TrimSpace(
		text[start:end],
	), end, true
}

// appendRawDeclaration preserves an unsupported or malformed declaration
// construct as raw CSS.
func appendRawDeclaration(
	declarations *[]Declaration,
	leadingComments []string,
	raw string,
) {
	raw = strings.TrimSpace(
		raw,
	)

	if raw == "" {
		return
	}

	*declarations = append(
		*declarations,
		Declaration{
			LeadingComment: leadingComments,

			Raw: []string{
				raw,
			},
		},
	)
}

// declarationValue returns the normalized value and the next parser index.
//
// A declaration terminated by ';' resumes immediately after the semicolon.
// An unterminated declaration consumes the remainder of the block.
func declarationValue(
	body string,
	valueStart int,
	valueEnd int,
) (string, int) {
	if valueEnd < 0 {
		return strings.TrimSpace(
			body[valueStart:],
		), len(body)
	}

	return strings.TrimSpace(
		body[valueStart : valueEnd+1],
	), valueEnd + 1
}

// findDeclarationColon finds the colon separating a declaration name from
// its value.
//
// Colons inside:
//
//   - strings
//   - comments
//   - parentheses
//   - attribute brackets
//
// are ignored.
func findDeclarationColon(
	text string,
	start int,
) int {
	state := parseState{}

	for index := start; index < len(text); index++ {
		advanceState(
			&state,
			text,
			&index,
		)

		if state.inComment ||
			state.inString {
			continue
		}

		if state.parens != 0 ||
			state.brackets != 0 {
			continue
		}

		switch text[index] {
		case ':':
			return index

		case ';',
			'{',
			'}':
			return -1
		}
	}

	return -1
}

// findDeclarationEnd finds the semicolon terminating a declaration.
//
// The closing brace is treated as the end of the declaration body and is
// not included in the returned index.
func findDeclarationEnd(
	text string,
	start int,
) int {
	state := parseState{}

	for index := start; index < len(text); index++ {
		advanceState(
			&state,
			text,
			&index,
		)

		if state.inComment ||
			state.inString {
			continue
		}

		if state.parens != 0 ||
			state.brackets != 0 {
			continue
		}

		switch text[index] {
		case ';':
			return index

		case '}':
			return index - 1
		}
	}

	return -1
}

// splitValueLines normalizes a declaration value into logical lines.
//
// Empty lines are removed because indentation and formatting are regenerated
// by the CSS writer.
func splitValueLines(
	value string,
) []string {
	value = normalizeLineEndings(
		value,
	)

	lines := strings.Split(
		value,
		"\n",
	)

	out := make(
		[]string,
		0,
		len(lines),
	)

	for _, line := range lines {
		line = strings.TrimSpace(
			line,
		)

		if line == "" {
			continue
		}

		out = append(
			out,
			line,
		)
	}

	if len(out) == 0 {
		return []string{""}
	}

	return out
}

// advanceState consumes the current character and updates the scanner state.
//
// When an escaped character occurs inside a string, the next character is
// consumed by incrementing the caller's loop index.
func advanceState(
	state *parseState,
	text string,
	index *int,
) {
	char := text[*index]

	if state.inComment {
		if char == '*' &&
			*index+1 < len(text) &&
			text[*index+1] == '/' {
			state.inComment = false
			(*index)++
		}

		return
	}

	if state.inString {
		if char == '\\' &&
			*index+1 < len(text) {
			(*index)++
			return
		}

		if char == state.quote {
			state.inString = false
		}

		return
	}

	if char == '/' &&
		*index+1 < len(text) &&
		text[*index+1] == '*' {
		state.inComment = true
		(*index)++
		return
	}

	if char == '"' ||
		char == '\'' {
		state.inString = true
		state.quote = char
		return
	}

	switch char {
	case '(':
		state.parens++

	case ')':
		if state.parens > 0 {
			state.parens--
		}

	case '[':
		state.brackets++

	case ']':
		if state.brackets > 0 {
			state.brackets--
		}
	}
}

// extractInlineComment returns the first top-level block comment in a
// declaration value.
//
// Comments inside strings, functions, or attribute expressions are ignored.
func extractInlineComment(
	value string,
) string {
	state := parseState{}

	for index := 0; index < len(value); index++ {
		if isTopLevelCommentStart(
			state,
			value,
			index,
		) {
			return strings.TrimSpace(
				value[index:],
			)
		}

		advanceState(
			&state,
			value,
			&index,
		)
	}

	return ""
}

// removeInlineComment removes the first top-level block comment from a
// declaration value.
func removeInlineComment(
	value string,
) string {
	state := parseState{}

	for index := 0; index < len(value); index++ {
		if isTopLevelCommentStart(
			state,
			value,
			index,
		) {
			return strings.TrimSpace(
				value[:index],
			)
		}

		advanceState(
			&state,
			value,
			&index,
		)
	}

	return strings.TrimSpace(
		value,
	)
}

// isTopLevelCommentStart reports whether a block comment starts at index
// outside strings, comments, parentheses, and attribute selectors.
func isTopLevelCommentStart(
	state parseState,
	text string,
	index int,
) bool {
	return !state.inComment &&
		!state.inString &&
		state.parens == 0 &&
		state.brackets == 0 &&
		isDeclarationCommentStart(
			text,
			index,
		)
}
