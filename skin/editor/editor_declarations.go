package editor

import (
	"strings"

	"github.com/zellydev-games/opensplit/skin/parser"
)

// formatRuleDeclarations converts parsed declarations back into the text
// representation used by the editor.
//
// Comments, raw declarations, multiline values, semicolons, and inline
// comments are preserved as far as the parser representation allows.
func FormatRuleDeclarations(
	declarations []parser.Declaration,
) string {
	var out strings.Builder

	for _, declaration := range declarations {
		for _, comment := range declaration.LeadingComment {
			out.WriteString(comment)
			out.WriteString("\n")
		}

		if declaration.Name == "" {
			for _, raw := range declaration.Raw {
				out.WriteString(raw)
				out.WriteString("\n")
			}

			continue
		}

		out.WriteString(declaration.Name)
		out.WriteString(": ")

		if len(declaration.Value) > 0 {
			out.WriteString(
				strings.Join(
					declaration.Value,
					"\n",
				),
			)

			lastValue := strings.TrimSpace(
				declaration.Value[len(declaration.Value)-1],
			)

			if !strings.HasSuffix(
				lastValue,
				";",
			) {
				out.WriteString(";")
			}
		}

		if declaration.InlineComment != "" {
			out.WriteString(" ")
			out.WriteString(
				declaration.InlineComment,
			)
		}

		out.WriteString("\n")
	}

	return strings.TrimSpace(
		out.String(),
	)
}
