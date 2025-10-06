//go:build linux && x11

package hotkeys

import "github.com/zellydev-games/opensplit/hotkeys/x11"

func SetupHotkeys() *x11.X11Manager {
	return x11.SetupHotkeys()
}
