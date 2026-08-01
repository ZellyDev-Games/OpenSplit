package parser

import (
	"strings"
)

// FormatCSS regenerates CSS formatting from the parsed rule tree.
//
// Formatting is intentionally normalized on save.
func FormatCSS(
	rules []Rule,
) string {

	var out strings.Builder

	for i, rule := range rules {

		if i > 0 {
			out.WriteString("\n\n")
		}

		writeRule(
			&out,
			rule,
			"",
		)
	}

	return strings.TrimSpace(
		out.String(),
	) + "\n"
}

func writeRule(
	out *strings.Builder,
	rule Rule,
	indent string,
) {

	// At-rules
	if rule.AtRule != "" {

		writeAtRule(
			out,
			rule,
			indent,
		)

		return
	}

	// Normal selector rule

	out.WriteString(indent)

	out.WriteString(
		strings.TrimSpace(
			rule.Selector,
		),
	)

	out.WriteString(
		" {\n",
	)

	writeDeclarations(
		out,
		indent+"    ",
		rule.Declarations,
	)

	out.WriteString(
		indent,
	)

	out.WriteString(
		"}",
	)
}

func writeAtRule(
	out *strings.Builder,
	rule Rule,
	indent string,
) {

	out.WriteString(
		indent,
	)

	out.WriteString(
		strings.TrimSpace(
			rule.AtRule,
		),
	)

	// @import "./file.css";
	//
	// @charset "utf8";
	//
	// @namespace ...
	if !rule.Block {

		if !strings.HasSuffix(
			strings.TrimSpace(rule.AtRule),
			";",
		) {

			out.WriteString(
				";",
			)
		}

		return
	}

	out.WriteString(
		" {\n",
	)

	// Some at-rules contain declarations:
	//
	// @font-face
	// @page
	if len(rule.Declarations) > 0 {

		writeDeclarations(
			out,
			indent+"    ",
			rule.Declarations,
		)
	}

	// Container at-rules:
	//
	// @layer
	// @media
	// @supports
	// @container
	// @keyframes
	for i, child := range rule.Children {

		if len(rule.Declarations) > 0 ||
			i > 0 {

			out.WriteString(
				"\n",
			)
		}

		writeRule(
			out,
			child,
			indent+"    ",
		)
	}

	out.WriteString(
		"\n",
	)

	out.WriteString(
		indent,
	)

	out.WriteString(
		"}",
	)
}

func writeDeclarations(
	out *strings.Builder,
	indent string,
	declarations []Declaration,
) {
	childIndent := indent + "    "

	for _, declaration := range declarations {
		for _, comment := range declaration.LeadingComment {
			out.WriteString(
				indent,
			)

			out.WriteString(
				comment,
			)

			out.WriteString(
				"\n",
			)
		}

		//
		// Raw CSS / unsupported constructs
		//
		if declaration.Name == "" {
			for _, raw := range declaration.Raw {
				out.WriteString(
					indent,
				)

				out.WriteString(
					raw,
				)

				out.WriteString(
					"\n",
				)
			}

			continue
		}

		out.WriteString(
			indent,
		)

		out.WriteString(
			declaration.Name,
		)

		out.WriteString(
			": ",
		)

		if len(declaration.Value) == 0 {

			out.WriteString(
				";\n",
			)

			continue
		}

		if len(declaration.Value) == 1 {
			out.WriteString(
				strings.TrimSpace(
					declaration.Value[0],
				),
			)

		} else {
			out.WriteString(
				"\n",
			)

			for i, line := range declaration.Value {
				out.WriteString(childIndent)

				out.WriteString(
					strings.TrimSpace(
						line,
					),
				)

				if i != len(declaration.Value)-1 {

					out.WriteString(
						"\n",
					)
				}
			}
		}

		if len(declaration.Value) > 0 {

			last :=
				strings.TrimSpace(
					declaration.Value[len(declaration.Value)-1],
				)

			if !strings.HasSuffix(
				last,
				";",
			) {
				out.WriteString(";")
			}
		}

		if declaration.InlineComment != "" {

			out.WriteString(
				" ",
			)

			out.WriteString(
				declaration.InlineComment,
			)
		}

		out.WriteString(
			"\n",
		)
	}
}
