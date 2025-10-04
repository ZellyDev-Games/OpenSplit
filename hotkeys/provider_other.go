//go:build !windows

package hotkeys

import "github.com/zellydev-games/opensplit/keyinfo"

type HotkeyProviderStub struct{}

func (h *HotkeyProviderStub) StartHook(func(data keyinfo.KeyData)) error {
	return nil
}

func (h *HotkeyProviderStub) Unhook() error {
	return nil
}

func SetupHotkeys() *HotkeyProviderStub {
	return &HotkeyProviderStub{}
}
