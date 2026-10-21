package serviceops

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/geodro/lerd/internal/config"
	"github.com/geodro/lerd/internal/plugins"
)

// A plugin takes the services installed from it, so nothing it brought keeps
// running; a service of the user's own on the same preset, and the data, stay.
func TestRemovePlugin_TakesTheServicesItShipped(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)
	rec := stubPodmanRemove(t)

	dir := filepath.Join(plugins.Dir(), "vmsvc")
	if err := os.MkdirAll(filepath.Join(dir, "services"), 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		plugins.ManifestFile:    "api_version: 1\nname: vmsvc\n",
		"services/vmcache.yaml": "name: vmcache\nimage: docker.io/library/redis:7.4\n",
	}
	for rel, body := range files {
		if err := os.WriteFile(filepath.Join(dir, rel), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, svc := range []*config.CustomService{
		{Name: "vmcache", Preset: "vmcache", Plugin: "vmsvc", Image: "docker.io/library/redis:7.4"},
		{Name: "mine", Preset: "vmcache", Image: "docker.io/library/redis:7.4"},
	} {
		if err := config.SaveCustomService(svc); err != nil {
			t.Fatal(err)
		}
	}
	data := mkTestDataDir(t, "vmcache", "kept")

	if err := RemovePlugin("vmsvc"); err != nil {
		t.Fatalf("RemovePlugin: %v", err)
	}
	if !slices.Contains(rec.stopped, "lerd-vmcache") || slices.Contains(rec.stopped, "lerd-mine") {
		t.Errorf("stopped %v, want only the shipped service", rec.stopped)
	}
	if _, err := config.LoadCustomService("vmcache"); err == nil {
		t.Error("the shipped service is still installed")
	}
	if _, err := config.LoadCustomService("mine"); err != nil {
		t.Errorf("the user's own service went too: %v", err)
	}
	if _, err := os.Stat(filepath.Join(data, "marker.txt")); err != nil {
		t.Errorf("the shipped service's data should be kept: %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("the plugin folder is still there")
	}
}

func TestRemovePlugin_RefusesABadName(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	if err := RemovePlugin("../services"); err == nil {
		t.Error("a name that climbs out of the plugins folder must be refused")
	}
}
