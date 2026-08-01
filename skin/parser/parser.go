package parser

import (
	"fmt"
	"os"
	"strings"
)

func ParseFile(
	filename string,
) ([]Rule, error) {

	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return Parse(
		filename,
		string(b),
	), nil
}

func Parse(
	filename string,
	css string,
) []Rule {

	rules :=
		parse(
			filename,
			css,
			1,
			"",
		)

	assignRuleIDs(
		rules,
		filename,
	)

	return rules
}

func assignRuleIDs(
	rules []Rule,
	filename string,
) {

	for i := range rules {

		rules[i].Order = i

		rules[i].ID =
			fmt.Sprintf(
				"%s:%d:%d",
				filename,
				rules[i].Line,
				i,
			)

		assignRuleIDs(
			rules[i].Children,
			filename,
		)
	}
}

func parse(
	filename string,
	css string,
	startLine int,
	layer string,
) []Rule {

	var rules []Rule

	i := 0
	line := startLine

	for i < len(css) {

		// Skip whitespace/comments.
		for i < len(css) {

			if css[i] == '\n' {
				line++
			}

			if strings.ContainsRune(
				" \t\r\n",
				rune(css[i]),
			) {
				i++
				continue
			}

			if i+1 < len(css) &&
				css[i] == '/' &&
				css[i+1] == '*' {

				end := strings.Index(
					css[i+2:],
					"*/",
				)

				if end < 0 {
					i = len(css)
					break
				}

				comment := css[i : i+end+4]

				line += strings.Count(
					comment,
					"\n",
				)

				i += end + 4
				continue
			}

			break
		}

		if i >= len(css) {
			break
		}

		start := i

		if css[i] == '@' {

			end := strings.IndexByte(
				css[i:],
				';',
			)

			brace := strings.IndexByte(
				css[i:],
				'{',
			)

			if end >= 0 &&
				(brace < 0 || end < brace) {

				end += i

				text :=
					strings.TrimSpace(
						css[i : end+1],
					)

				rules = append(
					rules,
					Rule{
						File: filename,

						Layer: layer,

						AtRule: text,

						Block: false,

						Prelude: strings.HasPrefix(
							strings.ToLower(text),
							"@import",
						),

						Line: line,
					},
				)

				line += strings.Count(
					text,
					"\n",
				)

				i = end + 1
				continue
			}
		}

		brace := strings.IndexByte(
			css[i:],
			'{',
		)

		if brace < 0 {
			break
		}

		brace += i

		header := strings.TrimSpace(
			css[start:brace],
		)

		depth := 1
		j := brace + 1

		for j < len(css) && depth > 0 {

			switch css[j] {

			case '{':
				depth++

			case '}':
				depth--
			}

			j++
		}

		if depth != 0 {
			break
		}

		body := css[brace+1 : j-1]

		headerLine := line

		line += strings.Count(
			css[start:j],
			"\n",
		)

		// Handle all block at-rules.
		//
		// Examples:
		//
		// @layer vars {}
		// @media (...) {}
		// @supports (...) {}
		// @font-face {}
		// @keyframes {}
		//
		if strings.HasPrefix(
			header,
			"@",
		) {

			if strings.HasPrefix(
				header,
				"@layer",
			) {

				name :=
					strings.TrimSpace(
						strings.TrimPrefix(
							header,
							"@layer",
						),
					)

				children :=
					parse(
						filename,
						body,
						headerLine+1,
						name,
					)

				rules =
					append(
						rules,
						Rule{

							File: filename,

							Layer: layer,

							AtRule: strings.TrimSpace(
								header,
							),

							Block: true,

							Children: children,

							Line: headerLine,
						},
					)

				i = j
				continue
			}

			var children []Rule
			var declarations []Declaration

			if isDeclarationAtRule(header) {

				declarations =
					ParseDeclarations(
						body,
					)

			} else {

				children =
					parse(
						filename,
						body,
						headerLine+1,
						layer,
					)
			}

			rules = append(
				rules,
				Rule{

					File: filename,

					Layer: layer,

					AtRule: strings.TrimSpace(
						header,
					),

					Block: true,

					Declarations: declarations,

					Children: children,

					Line: headerLine,
				},
			)

			i = j

			continue
		}

		if header == "" {
			i = j
			continue
		}

		selectors := splitSelectors(header)

		declarations := ParseDeclarations(
			strings.TrimSpace(body),
		)

		for _, selector := range selectors {

			selector = strings.Join(
				strings.Fields(selector),
				" ",
			)

			rules = append(
				rules,
				Rule{
					File:         filename,
					Layer:        layer,
					Selector:     selector,
					Declarations: cloneDeclarations(declarations),
					Line:         headerLine,
				},
			)
		}

		i = j
	}

	return rules
}

func isDeclarationAtRule(
	header string,
) bool {

	header =
		strings.ToLower(
			strings.TrimSpace(
				header,
			),
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

func cloneDeclarations(
	in []Declaration,
) []Declaration {

	out := make(
		[]Declaration,
		len(in),
	)

	for i, declaration := range in {

		out[i] = declaration

		out[i].Value =
			append(
				[]string{},
				declaration.Value...,
			)

		out[i].LeadingComment =
			append(
				[]string{},
				declaration.LeadingComment...,
			)

		out[i].Raw =
			append(
				[]string{},
				declaration.Raw...,
			)
	}

	return out
}

// splitSelectors splits a selector list while respecting
// (), [] and {} nesting.
func splitSelectors(
	header string,
) []string {

	var out []string

	var current strings.Builder

	parens := 0
	brackets := 0
	braces := 0

	for _, r := range header {

		switch r {

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

				s := strings.TrimSpace(current.String())

				if s != "" {
					out = append(
						out,
						s,
					)
				}

				current.Reset()
				continue
			}
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		s := strings.TrimSpace(current.String())

		if s != "" {
			out = append(
				out,
				s,
			)
		}
	}

	return out
}
