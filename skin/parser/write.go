package parser

import "strings"

// FormatCSS regenerates CSS from the parsed rule tree.
//
// Formatting is intentionally normalized when CSS is written. The parser
// retains the semantic rule structure while this package owns the concrete
// output formatting.
func FormatCSS(
	rules []Rule,
) string {
	rules = normalizeLayerRules(rules)

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

// normalizeLayerRules converts the editor's flat representation of layered
// rules back into nested @layer blocks.
//
// Rules loaded from disk already contain nested @layer nodes and are therefore
// preserved as-is. Rules created by the editor may instead be represented as
// ordinary rules carrying a Layer value. Consecutive rules belonging to the
// same layer are grouped into a generated @layer block.
//
// A layer boundary is created whenever:
//
//   - an existing @layer rule is encountered
//   - the layer name changes
//   - a rule has no layer
func normalizeLayerRules(
	rules []Rule,
) []Rule {
	if len(rules) == 0 {
		return nil
	}

	out := make(
		[]Rule,
		0,
		len(rules),
	)

	for i := 0; i < len(rules); i++ {
		rule := rules[i]

		// Existing @layer blocks already have the correct tree structure.
		if isLayerRule(rule) {
			out = append(
				out,
				rule,
			)

			continue
		}

		// Rules without a layer remain at the current scope.
		if rule.Layer == "" {
			out = append(
				out,
				rule,
			)

			continue
		}

		layer := rule.Layer

		children := make(
			[]Rule,
			0,
		)

		start := i

		for i < len(rules) {
			current := rules[i]

			if isLayerRule(current) ||
				current.Layer != layer {
				break
			}

			children = append(
				children,
				current,
			)

			i++
		}

		if len(children) == 0 {
			// This should not normally be reachable because the first rule
			// already has the requested layer, but retaining the guard keeps
			// this helper safe if its grouping conditions change later.
			i = start
			continue
		}

		// The loop exits with i pointing at the first rule that does not
		// belong to this layer. The outer loop will increment i again.
		i--

		out = append(
			out,
			Rule{
				File: filenameForLayer(
					children,
				),

				AtRule: "@layer " + layer,

				Block: true,

				Children: children,

				Line: children[0].Line,

				Order: children[0].Order,
			},
		)
	}

	return out
}

// isLayerRule reports whether a rule represents a block @layer rule.
func isLayerRule(
	rule Rule,
) bool {
	header := strings.TrimSpace(
		rule.AtRule,
	)

	return rule.Block &&
		strings.HasPrefix(
			strings.ToLower(header),
			"@layer",
		)
}

// filenameForLayer returns the source file associated with a generated layer
// block.
//
// Layer blocks are generated only for rules that originated from the same
// editor file, so the first child provides the appropriate filename.
func filenameForLayer(
	rules []Rule,
) string {
	if len(rules) == 0 {
		return ""
	}

	return rules[0].File
}

// writeRule writes one rule and its descendants.
//
// The rule can be either:
//
//   - a normal selector rule
//   - an at-rule
func writeRule(
	out *strings.Builder,
	rule Rule,
	indent string,
) {
	if rule.AtRule != "" {
		writeAtRule(
			out,
			rule,
			indent,
		)

		return
	}

	writeSelectorRule(
		out,
		rule,
		indent,
	)
}

// writeSelectorRule writes a normal CSS selector rule.
func writeSelectorRule(
	out *strings.Builder,
	rule Rule,
	indent string,
) {
	out.WriteString(indent)

	out.WriteString(
		strings.TrimSpace(
			rule.Selector,
		),
	)

	out.WriteString(" {\n")

	writeDeclarations(
		out,
		indent+"    ",
		rule.Declarations,
	)

	out.WriteString(indent)
	out.WriteString("}")
}

// writeAtRule writes both block and non-block at-rules.
//
// Examples of non-block at-rules:
//
//	@import "./file.css";
//	@charset "utf-8";
//	@namespace ...;
//
// Examples of block at-rules:
//
//	@layer
//	@media
//	@supports
//	@container
//	@font-face
//	@page
//	@keyframes
func writeAtRule(
	out *strings.Builder,
	rule Rule,
	indent string,
) {
	out.WriteString(indent)

	atRule := strings.TrimSpace(
		rule.AtRule,
	)

	out.WriteString(atRule)

	if !rule.Block {
		if !strings.HasSuffix(
			atRule,
			";",
		) {
			out.WriteString(";")
		}

		return
	}

	out.WriteString(" {\n")

	// Declaration-based at-rules such as @font-face and @page.
	if len(rule.Declarations) > 0 {
		writeDeclarations(
			out,
			indent+"    ",
			rule.Declarations,
		)
	}

	// Nested at-rules and selector rules.
	for i, child := range rule.Children {
		if len(rule.Declarations) > 0 || i > 0 {
			out.WriteString("\n")
		}

		writeRule(
			out,
			child,
			indent+"    ",
		)
	}

	out.WriteString("\n")
	out.WriteString(indent)
	out.WriteString("}")
}

// writeDeclarations writes declarations using normalized indentation.
//
// Declaration comments, multiline values, inline comments, and raw
// unsupported constructs are preserved by the parser's Declaration model.
func writeDeclarations(
	out *strings.Builder,
	indent string,
	declarations []Declaration,
) {
	childIndent := indent + "    "

	for _, declaration := range declarations {
		for _, comment := range declaration.LeadingComment {
			out.WriteString(indent)
			out.WriteString(comment)
			out.WriteString("\n")
		}

		// Raw CSS / unsupported constructs.
		if declaration.Name == "" {
			for _, raw := range declaration.Raw {
				out.WriteString(indent)
				out.WriteString(raw)
				out.WriteString("\n")
			}

			continue
		}

		out.WriteString(indent)
		out.WriteString(declaration.Name)
		out.WriteString(": ")

		if len(declaration.Value) == 0 {
			out.WriteString(";\n")
			continue
		}

		if len(declaration.Value) == 1 {
			out.WriteString(
				strings.TrimSpace(
					declaration.Value[0],
				),
			)
		} else {
			out.WriteString("\n")

			for i, line := range declaration.Value {
				out.WriteString(childIndent)
				out.WriteString(
					strings.TrimSpace(line),
				)

				if i < len(declaration.Value)-1 {
					out.WriteString("\n")
				}
			}
		}

		last := strings.TrimSpace(
			declaration.Value[len(declaration.Value)-1],
		)

		if !strings.HasSuffix(last, ";") {
			out.WriteString(";")
		}

		if declaration.InlineComment != "" {
			out.WriteString(" ")
			out.WriteString(declaration.InlineComment)
		}

		out.WriteString("\n")
	}
}
