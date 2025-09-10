//go:build !windows

package hotkeys

func SetupHotkeys() (HotkeyProvider, chan KeyInfo) {
	return nil, nil
}
