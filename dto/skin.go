package dto

// CSSFile represents one editable CSS file within a skin.
type CSSFile struct {
	Name string `json:"name"`

	Path string `json:"path"`

	Contents string `json:"contents"`

	// Original contents used for parser byte offsets.
	//
	// Parser.Rule.BodyStart/BodyEnd refer to this string.
	// It must not change until the editor is reloaded.
	OriginalContents string `json:"originalContents"`

	Text bool `json:"text"`

	URL string `json:"url"`

	Type string `json:"type"`
}

// CSSRule represents a parsed CSS rule.
//
// Rules form a tree.
//
// Examples:
//
// @layer skins {}
// @media (...) {}
// @supports (...) {}
// @font-face {}
//
// are represented as rules with AtRule populated and Children containing
// nested rules.
type CSSRule struct {
	ID string `json:"id"`

	File string `json:"file"`

	Layer string `json:"layer"`

	Selector string `json:"selector"`

	Body string `json:"body"`

	// CSS at-rule text.
	//
	// Examples:
	//
	// @layer skins
	// @media (max-width:600px)
	// @supports(display:grid)
	// @font-face
	// @import "./vars.css";
	//
	AtRule string `json:"atRule"`

	// Whether AtRule owns a block.
	//
	// true:
	//   @media (...) {}
	//
	// false:
	//   @import "./vars.css";
	Block bool `json:"block"`

	// Child rules contained by this rule.
	Children []CSSRule `json:"children"`

	// True for rules that must remain before normal CSS.
	//
	// Example:
	//
	// @import "./vars.css";
	//
	Prelude bool `json:"prelude"`

	Line int `json:"line"`

	Order int `json:"order"`

	// Parent parser rule.
	ParentID string `json:"parentId"`

	// Create indicates this is a new rule.
	Create bool `json:"create"`

	// Delete indicates this rule should be removed.
	Delete bool `json:"delete"`
}

type CSSRuleEditor struct {
	ID string `json:"id"`

	File string `json:"file"`

	Selector string `json:"selector"`

	Layer string `json:"layer"`

	// Parent at-rule/rule context.
	//
	// Used when creating or moving rules.
	ParentID string `json:"parentId"`

	ParentAtRule string `json:"parentAtRule"`

	Body string `json:"body"`

	Create bool `json:"create"`

	Delete bool `json:"delete"`

	OriginalFile string `json:"originalFile"`

	OriginalID string `json:"originalId"`
}

// SkinElement represents an editable UI element in the skin editor.
//
// Elements are frontend targets. The editor uses the selector to locate
// matching CSS rules across files.
type SkinElement struct {
	// Stable identifier used by the frontend.
	ID string `json:"id"`

	// Human readable name shown in the element picker.
	Label string `json:"label"`

	// CSS selector associated with this element.
	Selector string `json:"selector"`

	// Optional description shown in the editor.
	Description string `json:"description,omitempty"`

	File   string `json:"file,omitempty"`
	Line   int    `json:"line,omitempty"`
	Layer  string `json:"layer,omitempty"`
	RuleID string `json:"ruleId,omitempty"`
}

// SkinEditorTarget is the current editor selection.
//
// This allows the backend to own editor navigation state so the
// frontend only renders the current selection.
type SkinEditorTarget struct {
	ElementID string `json:"elementId"`

	File string `json:"file"`

	Selector string `json:"selector"`

	RuleID string `json:"ruleId"`

	ParentID string `json:"parentId"`

	Mode string `json:"mode"`
}

type SkinPreviewElement struct {
	ID string `json:"id"`

	Label string `json:"label"`

	Selector string `json:"selector"`
}
