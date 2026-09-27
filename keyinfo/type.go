package keyinfo

// KeyData is the Go-friendly struct to capture key code and key name data from the OS
type KeyData struct {
	KeyCode             int      `json:"key_code"`
	LocaleName          string   `json:"locale_name"`
	Modifiers           []string `json:"modifiers"`
	ModifierLocaleNames []string `json:"modifier_locale_names"`
}

// NewKeyData constructs a normalized KeyData value.
func NewKeyData(
	keyCode int,
	localeName string,
	modifiers []string,
	modifierLocaleNames []string,
) KeyData {
	if modifiers == nil {
		modifiers = []string{}
	}

	if modifierLocaleNames == nil {
		modifierLocaleNames = []string{}
	}

	return KeyData{
		KeyCode:             keyCode,
		LocaleName:          localeName,
		Modifiers:           modifiers,
		ModifierLocaleNames: modifierLocaleNames,
	}
}
