package config

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/geodro/lerd/internal/platform"
	"github.com/geodro/lerd/internal/wsl"
)

// WindowsThemeID is the id the Windows desktop theme takes in the picker, the
// same as the built-in Windows palette it stands in for.
const WindowsThemeID = "windows"

// windowsDesktop follows the Windows accent from inside WSL, the one place lerd
// runs with Windows as its desktop. The accent lives in the registry, which has
// no file to watch, so the dashboard picks a change up on its next load.
func windowsDesktop() Desktop {
	if !platform.Current.WindowsSide {
		return Desktop{}
	}
	out := desktopToolOutput(wsl.WindowsTool("reg.exe"), "query", `HKCU\Software\Microsoft\Windows\CurrentVersion\Explorer\Accent`, "/v", "AccentPalette")
	return Desktop{Theme: windowsTheme(parseRegBinary(out, "AccentPalette"))}
}

// windowsTheme turns Windows' AccentPalette into a theme with the shades WinUI
// uses, Dark1 and Dark2 in light mode, Light2 and Light3 in dark; the surfaces
// are not the user's choice and the built-in Windows palette has them.
func windowsTheme(palette string) *UITheme {
	b, err := hex.DecodeString(palette)
	if err != nil || len(b) < 28 {
		return nil
	}
	shade := func(i int) string { return fmt.Sprintf("#%02x%02x%02x", b[i*4], b[i*4+1], b[i*4+2]) }
	return &UITheme{
		ID:              WindowsThemeID,
		Name:            "Windows",
		Accent:          shade(4),
		AccentHover:     shade(5),
		AccentDark:      shade(1),
		AccentHoverDark: shade(0),
		Source:          UIThemeSourceDesktop,
	}
}

// parseRegBinary pulls a REG_BINARY value's hex out of `reg.exe query` output.
func parseRegBinary(out, name string) string {
	for _, line := range strings.Split(out, "\n") {
		f := strings.Fields(line)
		if len(f) == 3 && f[0] == name && f[1] == "REG_BINARY" {
			return f[2]
		}
	}
	return ""
}
