//go:build linux && x11

package hotkeys

import "github.com/zellydev-games/opensplit/hotkeys/x11"

// SetupHotkeys wraps the C implementation in the x11 package and exposes it to the system at large.
//
// This indirection is here to keep the c files out of the hotkeys package so that CGO doesn't have to be enabled
// on a platform that doesn't need it (e.g. Windows)
func SetupHotkeys() *x11.Manager {
	return x11.SetupHotkeys()
}
