package parser

import (
	"strings"
	"unicode"
)

type parseState struct {
	parens   int
	brackets int

	inString bool
	quote    byte

	inComment bool
}

func ParseDeclarations(
	body string,
) []Declaration {

	body = strings.ReplaceAll(
		body,
		"\r\n",
		"\n",
	)

	var declarations []Declaration
	var pendingComments []string

	i := 0

	for i < len(body) {

		// Skip whitespace between declarations.
		for i < len(body) {

			switch body[i] {

			case ' ', '\t', '\r', '\n':
				i++

			default:
				goto begin
			}
		}

	begin:

		if i >= len(body) {
			break
		}

		// Standalone comment.
		if strings.HasPrefix(body[i:], "/*") {

			end := strings.Index(
				body[i+2:],
				"*/",
			)

			if end < 0 {

				pendingComments = append(
					pendingComments,
					strings.TrimSpace(body[i:]),
				)

				break
			}

			end += i + 4

			pendingComments = append(
				pendingComments,
				strings.TrimSpace(body[i:end]),
			)

			i = end
			continue
		}

		nameStart := i

		colon := findDeclarationColon(
			body,
			i,
		)

		if colon < 0 {

			raw := strings.TrimSpace(
				body[nameStart:],
			)

			if raw != "" {

				declarations = append(
					declarations,
					Declaration{
						LeadingComment: pendingComments,
						Raw:            []string{raw},
					},
				)
			}

			break
		}

		name := strings.TrimSpace(
			body[nameStart:colon],
		)

		if name == "" {

			i = colon + 1
			continue
		}

		valueStart := colon + 1

		valueEnd := findDeclarationEnd(
			body,
			valueStart,
		)

		var value string

		if valueEnd < 0 {

			value = strings.TrimSpace(
				body[valueStart:],
			)

			i = len(body)

		} else {

			value = strings.TrimSpace(
				body[valueStart : valueEnd+1],
			)

			i = valueEnd + 1
		}

		inline := extractInlineComment(
			value,
		)

		value = removeInlineComment(
			value,
		)

		declarations = append(
			declarations,
			Declaration{
				Name:           name,
				Value:          splitValueLines(value),
				InlineComment:  inline,
				LeadingComment: pendingComments,
			},
		)

		pendingComments = nil
	}

	return declarations
}

func findDeclarationColon(
	text string,
	start int,
) int {

	state := parseState{}

	for i := start; i < len(text); i++ {

		advanceState(
			&state,
			text,
			&i,
		)

		if state.inComment ||
			state.inString {

			continue
		}

		if state.parens != 0 ||
			state.brackets != 0 {

			continue
		}

		switch text[i] {

		case ':':
			return i

		case ';':
			// malformed declaration
			return -1

		case '{',
			'}':
			return -1
		}
	}

	return -1
}

func findDeclarationEnd(
	text string,
	start int,
) int {

	state := parseState{}

	for i := start; i < len(text); i++ {

		advanceState(
			&state,
			text,
			&i,
		)

		if state.inComment ||
			state.inString {

			continue
		}

		if state.parens != 0 ||
			state.brackets != 0 {

			continue
		}

		switch text[i] {

		case ';':
			return i

		case '}':
			return i - 1
		}
	}

	return -1
}

func splitValueLines(
	value string,
) []string {

	value = strings.ReplaceAll(
		value,
		"\r\n",
		"\n",
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

		line = strings.TrimRightFunc(
			line,
			unicode.IsSpace,
		)

		line = strings.TrimLeftFunc(
			line,
			unicode.IsSpace,
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

func advanceState(
	state *parseState,
	text string,
	i *int,
) {

	c := text[*i]

	if state.inComment {

		if c == '*' &&
			*i+1 < len(text) &&
			text[*i+1] == '/' {

			state.inComment = false
			*i++
		}

		return
	}

	if state.inString {

		if c == '\\' &&
			*i+1 < len(text) {

			*i++
			return
		}

		if c == state.quote {
			state.inString = false
		}

		return
	}

	if c == '/' &&
		*i+1 < len(text) &&
		text[*i+1] == '*' {

		state.inComment = true
		*i++
		return
	}

	if c == '"' ||
		c == '\'' {

		state.inString = true
		state.quote = c
		return
	}

	switch c {

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

func extractInlineComment(
	value string,
) string {

	state := parseState{}

	for i := 0; i < len(value); i++ {

		if !state.inComment &&
			!state.inString &&
			value[i] == '/' &&
			i+1 < len(value) &&
			value[i+1] == '*' {

			return strings.TrimSpace(value[i:])
		}

		advanceState(
			&state,
			value,
			&i,
		)
	}

	return ""
}

func removeInlineComment(
	value string,
) string {

	state := parseState{}

	for i := 0; i < len(value); i++ {

		if !state.inComment &&
			!state.inString &&
			value[i] == '/' &&
			i+1 < len(value) &&
			value[i+1] == '*' {

			return strings.TrimSpace(value[:i])
		}

		advanceState(
			&state,
			value,
			&i,
		)
	}

	return strings.TrimSpace(value)
}
