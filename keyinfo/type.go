package keyinfo

// KeyData is the Go-friendly struct to capture key code and key name data from the OS
type KeyData struct {
	KeyCode            int      `json:"key_code"`
	LocaleName         string   `json:"locale_name"`
	Modifiers          []int    `json:"modifiers"`
	ModifierLocalNames []string `json:"modifier_locale_names"`
}

func NewKeyData(kCode int, localeName string, modifiers []int, modifierLocalNames []string) KeyData {
	if modifiers == nil {
		modifiers = []int{}
	}

	if modifierLocalNames == nil {
		modifierLocalNames = []string{}
	}

	return KeyData{
		KeyCode:            kCode,
		LocaleName:         localeName,
		Modifiers:          modifiers,
		ModifierLocalNames: modifierLocalNames,
	}
}
