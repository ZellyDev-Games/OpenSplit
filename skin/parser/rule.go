package parser

type Rule struct {
	ID string `json:"id"`

	File string `json:"file"`

	Layer string `json:"layer"`

	Selector string `json:"selector"`

	Declarations []Declaration `json:"declarations"`

	AtRule string `json:"atRule"`

	Block bool `json:"block"`

	Children []Rule `json:"children"`

	Line int `json:"line"`

	Order int `json:"order"`

	ParentID string `json:"parentId"`

	Prelude bool `json:"prelude"`
}

type Declaration struct {
	// Empty Name indicates this is not a property declaration.
	// It may instead represent a standalone comment or raw CSS.
	Name string `json:"name"`

	// Value is split into logical lines.
	//
	// Single-line:
	//
	//     color: red;
	//
	// becomes
	//
	//     []string{"red;"}
	//
	// Multiline:
	//
	//     background:
	//         linear-gradient(
	//             red,
	//             blue
	//         );
	//
	// becomes
	//
	//     []string{
	//         "linear-gradient(",
	//         "red,",
	//         "blue",
	//         ");",
	//     }
	Value []string `json:"value"`

	// Inline comment after the declaration.
	//
	// color: red; /* comment */
	InlineComment string `json:"inlineComment"`

	// Standalone comment immediately preceding the declaration.
	LeadingComment []string `json:"leadingComment"`

	// Raw preserves constructs that are not CSS declarations
	// (future-proofing for @apply, nested syntax, etc.).
	Raw []string `json:"raw"`
}

type File struct {
	Name string `json:"name"`

	Layer string `json:"layer"`

	Rules []Rule `json:"rules"`
}
