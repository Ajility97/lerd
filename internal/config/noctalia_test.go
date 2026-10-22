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
	stubNoctaliaRunning(t, true)
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

func stubNoctaliaRunning(t *testing.T, running bool) {
	t.Helper()
	prev := noctaliaRunning
	noctaliaRunning = func() bool { return running }
	t.Cleanup(func() { noctaliaRunning = prev })
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
}

// The watch hangs off the registration alone, since at login Noctalia renders
// its file after lerd-ui has started.
func TestNoctaliaWatchFollowsTheRegistration(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	if dir, _ := NoctaliaWatch(); dir != "" {
		t.Errorf("NoctaliaWatch() = %q with no registration, want none", dir)
	}
	if err := os.MkdirAll(filepath.Dir(NoctaliaRegistrationPath()), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(NoctaliaRegistrationPath(), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	dir, names := NoctaliaWatch()
	if dir != filepath.Dir(NoctaliaThemePath()) || len(names) != 1 || names[0] != "theme.yaml" {
		t.Errorf("NoctaliaWatch() = %q, %v, want the rendered file", dir, names)
	}
}

// Back in Plasma, the file Noctalia rendered in an earlier session is still on
// disk, and Plasma's entry, focus tone included, has to win again.
func TestCurrentDesktopIgnoresNoctaliaThatIsNotRunning(t *testing.T) {
	writeKdeGlobals(t, "KDE", breezeDark)
	writeNoctaliaTheme(t, noctaliaRendered, "")
	stubNoctaliaRunning(t, false)
	d := CurrentDesktop()
	if d.Theme == nil || d.Theme.ID != PlasmaThemeID {
		t.Fatalf("CurrentDesktop() = %+v, want the Plasma theme", d.Theme)
	}
}
