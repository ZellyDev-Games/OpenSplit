//go:build darwin

package hotkeys

import "github.com/zellydev-games/opensplit/hotkeys/darwin"

func SetupHotkeys() *darwin.Manager {
	return new(darwin.Manager)
}
