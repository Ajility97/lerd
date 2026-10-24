package config

import (
	"io/fs"
	"testing"

	"gopkg.in/yaml.v3"
)

// Every preset lerd ships names itself, the same rule the store's schema guard
// holds its presets to, so no name is left to be guessed from the lowercase one.
func TestEmbeddedPresets_CarryALabel(t *testing.T) {
	paths, err := fs.Glob(presetFS, "presets/*.yaml")
	if err != nil || len(paths) == 0 {
		t.Fatalf("no embedded presets: %v", err)
	}
	for _, path := range paths {
		data, err := presetFS.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		var p Preset
		if err := yaml.Unmarshal(data, &p); err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		if p.Label == "" {
			t.Errorf("%s has no label", path)
		}
	}
}
