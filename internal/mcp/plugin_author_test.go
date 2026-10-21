package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/geodro/lerd/internal/plugins"
)

func callPlugin(t *testing.T, args map[string]any) (string, bool) {
	t.Helper()
	res, rpcErr := groupDispatch["plugin"][args["action"].(string)](args)
	if rpcErr != nil {
		t.Fatalf("rpc error: %+v", rpcErr)
	}
	m := res.(map[string]any)
	isErr, _ := m["isError"].(bool)
	return toolText(res), isErr
}

// An assistant scaffolds a plugin, sees it load with every surface, breaks it,
// and sees why it no longer loads.
func TestPluginTool_NewThenList(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	text, isErr := callPlugin(t, map[string]any{"action": "new", "name": "notes"})
	if isErr || !strings.Contains(text, filepath.Join(plugins.Dir(), "notes")) {
		t.Fatalf("new = %q (error %v), want the plugin's path", text, isErr)
	}
	if _, isErr := callPlugin(t, map[string]any{"action": "new", "name": "notes"}); !isErr {
		t.Error("a second new over the same name must fail")
	}

	var got struct {
		Installed []struct {
			Name     string   `json:"name"`
			Surfaces []string `json:"surfaces"`
		} `json:"installed"`
		Broken []struct{ Error string } `json:"broken"`
	}
	text, _ = callPlugin(t, map[string]any{"action": "list"})
	if err := json.Unmarshal([]byte(text), &got); err != nil {
		t.Fatalf("%v: %s", err, text)
	}
	if len(got.Installed) != 1 || got.Installed[0].Name != "notes" || len(got.Installed[0].Surfaces) != 9 {
		t.Fatalf("list = %s, want notes with all nine surfaces", text)
	}

	manifest := filepath.Join(plugins.Dir(), "notes", plugins.ManifestFile)
	if err := os.WriteFile(manifest, []byte("api_version: 1\nname: notes\nhooks:\n  site.renamed: bin/greet\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	text, _ = callPlugin(t, map[string]any{"action": "list"})
	got.Installed, got.Broken = nil, nil
	_ = json.Unmarshal([]byte(text), &got)
	if len(got.Installed) != 0 || len(got.Broken) != 1 || !strings.Contains(got.Broken[0].Error, "site.renamed") {
		t.Errorf("list = %s, want notes broken naming the unknown event", text)
	}
}

func TestPluginTool_InstallTakesOnlyALocalFolder(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	src := filepath.Join(t.TempDir(), "notes")
	if err := plugins.Scaffold("notes", src); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"https://github.com/someone/notes", "notes", ""} {
		if _, isErr := callPlugin(t, map[string]any{"action": "install", "path": path}); !isErr {
			t.Errorf("install %q must be refused", path)
		}
	}
	if text, isErr := callPlugin(t, map[string]any{"action": "install", "path": src}); isErr {
		t.Fatalf("install from a folder: %s", text)
	}
	if _, err := plugins.Find("notes"); err != nil {
		t.Errorf("installed plugin does not load: %v", err)
	}
}

func TestPluginTool_GuideIsTheEmbeddedGuide(t *testing.T) {
	text, _ := callPlugin(t, map[string]any{"action": "guide"})
	if text != plugins.Guide || !strings.Contains(text, "plugin.yaml") {
		t.Error("guide must return the plugin guide")
	}
}
