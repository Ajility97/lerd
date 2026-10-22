package config

import (
	"os"
	"path/filepath"
	"testing"
)

// noctaliaRendered is a file as the lerd template renders it, with Noctalia's
// Eldritch tones.
const noctaliaRendered = `accent: "#fb5bb6"
accent_dark: "#37f499"
bg: "#212337"
card: "#292e42"
border: "#3b4261"
muted: "#abb4da"
bg_light: "#d5d9db"
card_light: "#f0f3f4"
border_light: "#c9cbcd"
`

// writeNoctaliaTheme lays out the file Noctalia renders the lerd template into
// and the settings it keeps the palette choice in.
func writeNoctaliaTheme(t *testing.T, rendered, settings string) {
	t.Helper()
	state := t.TempDir()
	t.Setenv("XDG_STATE_HOME", state)
	for path, data := range map[string]string{
		filepath.Join(state, "lerd", "noctalia", "theme.yaml"): rendered,
		filepath.Join(state, "noctalia", "settings.toml"):      settings,
	} {
		if data == "" {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestNoctaliaThemeReadsTheRenderedPalette(t *testing.T) {
	writeNoctaliaTheme(t, noctaliaRendered, "[theme]\nsource = \"builtin\"\nbuiltin = \"Eldritch\"\n")
	theme := NoctaliaTheme()
	if theme == nil {
		t.Fatal("NoctaliaTheme() = nil, want the rendered palette")
	}
	want := UITheme{
		ID: NoctaliaThemeID, Name: "Noctalia (Eldritch)", Source: UIThemeSourceDesktop,
		Accent: "#fb5bb6", AccentDark: "#37f499",
		Bg: "#212337", Card: "#292e42", Border: "#3b4261", Muted: "#abb4da",
		BgLight: "#d5d9db", CardLight: "#f0f3f4", BorderLight: "#c9cbcd",
	}
	if *theme != want {
		t.Errorf("NoctaliaTheme() = %+v, want %+v", *theme, want)
	}
}

func TestNoctaliaThemeNamesThePalette(t *testing.T) {
	cases := map[string]string{
		"[theme]\nsource = \"community\"\nbuiltin = \"Eldritch\"\ncommunity_palette = \"Oxocarbon\"\n": "Noctalia (Oxocarbon)",
		"[theme]\nsource = \"wallpaper\"\nbuiltin = \"Eldritch\"\n":                                    "Noctalia (wallpaper)",
		"[theme]\nsource = \"custom\"\n":                                                               "Noctalia",
		"":                                                                                             "Noctalia",
	}
	for settings, want := range cases {
		writeNoctaliaTheme(t, noctaliaRendered, settings)
		if got := NoctaliaTheme().Name; got != want {
			t.Errorf("settings %q: name = %q, want %q", settings, got, want)
		}
	}
}

// With no template rendered there is nothing to follow, even where Noctalia
// itself is installed.
func TestNoctaliaThemeWithoutTheTemplate(t *testing.T) {
	writeNoctaliaTheme(t, "", "[theme]\nsource = \"builtin\"\nbuiltin = \"Eldritch\"\n")
	if theme := NoctaliaTheme(); theme != nil {
		t.Errorf("NoctaliaTheme() = %+v, want nil", theme)
	}
}

func TestNoctaliaThemeWithoutAnAccent(t *testing.T) {
	writeNoctaliaTheme(t, "bg: \"#212337\"\n", "")
	if theme := NoctaliaTheme(); theme != nil {
		t.Errorf("NoctaliaTheme() = %+v, want nil", theme)
	}
}

// Noctalia runs on niri and Hyprland next to a kdeglobals any Qt app leaves
// behind, and is the shell actually painting the desktop.
func TestCurrentDesktopPrefersNoctaliaOverPlasma(t *testing.T) {
	writeKdeGlobals(t, "KDE", breezeDark)
	writeNoctaliaTheme(t, noctaliaRendered, "")
	d := CurrentDesktop()
	if d.Theme == nil || d.Theme.ID != NoctaliaThemeID {
		t.Fatalf("CurrentDesktop() = %+v, want the Noctalia theme", d.Theme)
	}
	if d.WatchDir != filepath.Dir(NoctaliaThemePath()) {
		t.Errorf("WatchDir = %q, want %q", d.WatchDir, filepath.Dir(NoctaliaThemePath()))
	}
}
