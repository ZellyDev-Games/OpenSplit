package skin

import (
	"github.com/zellydev-games/opensplit/dto"
)

// SkinModel is the complete backend supplied skin editor state.
//
// The backend owns:
//   - selected element
//   - selected file
//   - selected rule
//   - editor mode
//
// The frontend only renders this model.
type SkinModel struct {
	// Active skin.
	Name string `json:"name"`

	// Installed skins.
	Skins []string `json:"skins"`

	// Absolute directory.
	Directory string `json:"directory"`

	// Stylesheet URL.
	StyleSheet string `json:"styleSheet"`

	// Editable files.
	Files []dto.CSSFile `json:"files"`

	// Preview targets.
	Elements []dto.SkinElement `json:"elements"`

	// Parsed rules.
	Rules []dto.CSSRule `json:"rules"`

	Target dto.SkinEditorTarget `json:"target"`
	// activeRule: CSSRuleEditor | null;

	ActiveRule *dto.CSSRuleEditor `json:"activeRule"`

	Dirty bool `json:"dirty"`

	Revision uint64 `json:"revision"`
}
