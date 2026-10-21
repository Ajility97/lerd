package plugins

import (
	"fmt"
	"runtime"
	"slices"
	"strings"
)

// platforms are the values `platforms:` takes. Windows is not one: lerd runs
// inside WSL there, so its plugins run on linux.
var platforms = map[string]bool{"linux": true, "macos": true}

// runningPlatform is the platform a plugin is judged against. A var so tests
// can pretend to be the other one.
var runningPlatform = platformOf(runtime.GOOS)

func platformOf(goos string) string {
	if goos == "darwin" {
		return "macos"
	}
	return "linux"
}

// RunsHere reports whether a plugin listing these platforms runs on this
// machine; a plugin that lists none runs everywhere lerd does.
func RunsHere(list []string) bool {
	return len(list) == 0 || slices.Contains(list, runningPlatform)
}

func (p *Plugin) checkPlatforms() error {
	for _, pl := range p.Platforms {
		if !platforms[pl] {
			return fmt.Errorf("unknown platform %q: use linux or macos", pl)
		}
	}
	if !RunsHere(p.Platforms) {
		return fmt.Errorf("runs only on %s (this is %s)", strings.Join(p.Platforms, ", "), runningPlatform)
	}
	return nil
}
