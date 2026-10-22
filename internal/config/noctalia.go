package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// NoctaliaThemeID is the id the Noctalia desktop theme takes in the picker. It
// is reserved: a user theme file of the same name would be shadowed by it.
const NoctaliaThemeID = "noctalia"

// noctaliaDesktop reads Noctalia as a desktop to follow. Noctalia keeps no
// colours file of its own; it renders user templates on every palette change,
// and the lerd template writes its file in place, so the watch sits on its
// directory and filters for the name.
func noctaliaDesktop() Desktop {
	theme := NoctaliaTheme()
	if theme == nil {
		return Desktop{}
	}
	return Desktop{
		Theme:      theme,
		WatchDir:   filepath.Dir(NoctaliaThemePath()),
		WatchNames: []string{filepath.Base(NoctaliaThemePath())},
	}
}

// NoctaliaThemePath is where the lerd template, registered as a Noctalia user
// template, renders the palette Noctalia is on.
func NoctaliaThemePath() string {
	return filepath.Join(xdgStateHome(), "lerd", "noctalia", "theme.yaml")
}

// NoctaliaTemplatePath is the lerd template Noctalia renders into
// NoctaliaThemePath.
func NoctaliaTemplatePath() string {
	return filepath.Join(ConfigDir(), "noctalia", "lerd.yaml")
}

// NoctaliaRegistrationPath registers the lerd template with Noctalia. It is a
// file of its own in Noctalia's config directory, which Noctalia reads beside
// settings.toml, so the settings its GUI manages stay untouched.
func NoctaliaRegistrationPath() string {
	return filepath.Join(xdgConfigHome(), "noctalia", "lerd.toml")
}

// NoctaliaTheme reads the palette the lerd template rendered as a dashboard
// theme, or returns nil where there is none. Anything unreadable is absence
// rather than an error, as with Omarchy: Noctalia rewrites the file on the next
// palette change, and there is nothing in it the user could correct.
func NoctaliaTheme() *UITheme {
	// The rendered file outlives the session, so on a machine that also runs
	// Plasma it would shadow Plasma's entry long after Noctalia was left.
	if !noctaliaRunning() {
		return nil
	}
	data, err := os.ReadFile(NoctaliaThemePath())
	if err != nil {
		return nil
	}
	theme, err := decodeUITheme(data)
	if err != nil || theme.Accent == "" {
		return nil
	}
	theme.ID = NoctaliaThemeID
	theme.Name = noctaliaThemeName()
	theme.Source = UIThemeSourceDesktop
	return theme
}

// noctaliaRunning says whether the shell is up, by its process name. It is a
// variable so tests can stand in for a session.
var noctaliaRunning = func() bool {
	comms, _ := filepath.Glob("/proc/[0-9]*/comm")
	for _, path := range comms {
		if data, err := os.ReadFile(path); err == nil && string(data) == "noctalia\n" {
			return true
		}
	}
	return false
}

// noctaliaThemeName labels the entry with the palette Noctalia is on. The
// template has no token for it, so it comes from Noctalia's own settings, where
// a custom scheme's name is not recorded.
func noctaliaThemeName() string {
	var s struct {
		Theme struct {
			Source           string `toml:"source"`
			Builtin          string `toml:"builtin"`
			CommunityPalette string `toml:"community_palette"`
		} `toml:"theme"`
	}
	data, _ := os.ReadFile(filepath.Join(xdgStateHome(), "noctalia", "settings.toml"))
	_ = toml.Unmarshal(data, &s)
	name := map[string]string{
		"builtin":   s.Theme.Builtin,
		"community": s.Theme.CommunityPalette,
		"wallpaper": "wallpaper",
	}[s.Theme.Source]
	if name == "" {
		return "Noctalia"
	}
	return "Noctalia (" + name + ")"
}
