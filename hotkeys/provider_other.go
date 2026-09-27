//go:build !windows && !wayland && !x11 && !darwin

package hotkeys

import (
	"context"

	"github.com/zellydev-games/opensplit/command"
	"github.com/zellydev-games/opensplit/keyinfo"
)

type HotkeyProviderStub struct{}

func (h *HotkeyProviderStub) Start(context.Context) error {
	return nil
}

func (h *HotkeyProviderStub) Configure(
	map[command.Command]keyinfo.KeyData,
) error {
	return nil
}

func (h *HotkeyProviderStub) Enable() error {
	return nil
}

func (h *HotkeyProviderStub) Disable() error {
	return nil
}

func (h *HotkeyProviderStub) SetCommandCallback(
	func(command.Command),
) {
}

func (h *HotkeyProviderStub) Close() error {
	return nil
}

func SetupHotkeys() *HotkeyProviderStub {
	return &HotkeyProviderStub{}
}
