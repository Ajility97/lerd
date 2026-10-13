package config

import (
	"os"
	"path/filepath"
	"testing"
)

// writeOverlay puts a user overlay for name in the sandboxed frameworks dir.
func writeOverlay(t *testing.T, name, body string) {
	t.Helper()
	if err := os.MkdirAll(FrameworksDir(), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(FrameworksDir(), name+".yaml"), []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
}

// The overlay is the user's own file, so its nginx block reaches the site the
// way its workers do, rather than being parsed and dropped.
func TestGetFramework_OverlayNginxReplacesTheBase(t *testing.T) {
	setConfigDir(t)
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	writeOverlay(t, "laravel", "name: laravel\nnginx:\n  snippet: |\n    location /vm-overlay { return 204; }\n  fastcgi_params:\n    VM_OVERLAY: \"{{root}}\"\n")

	fw, ok := GetFramework("laravel")
	if !ok {
		t.Fatal("laravel did not resolve")
	}
	if fw.Nginx == nil {
		t.Fatal("the overlay's nginx block was dropped")
	}
	if fw.Nginx.Snippet != "location /vm-overlay { return 204; }\n" {
		t.Errorf("snippet = %q, want the overlay's", fw.Nginx.Snippet)
	}
	if fw.Nginx.FastCGIParams["VM_OVERLAY"] != "{{root}}" {
		t.Errorf("fastcgi_params = %v, want the overlay's", fw.Nginx.FastCGIParams)
	}
}

// An overlay that only adds workers leaves the base's nginx block alone.
func TestGetFramework_OverlayWithoutNginxKeepsTheBase(t *testing.T) {
	setConfigDir(t)
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	if err := os.MkdirAll(StoreFrameworksDir(), 0755); err != nil {
		t.Fatal(err)
	}
	store := "name: acme\nnginx:\n  snippet: |\n    location /static { try_files $uri =404; }\n"
	if err := os.WriteFile(filepath.Join(StoreFrameworksDir(), "acme.yaml"), []byte(store), 0644); err != nil {
		t.Fatal(err)
	}
	writeOverlay(t, "acme", "name: acme\nworkers:\n  pulse:\n    command: php acme pulse\n")

	fw, ok := GetFramework("acme")
	if !ok {
		t.Fatal("acme did not resolve")
	}
	if fw.Nginx == nil || fw.Nginx.Snippet != "location /static { try_files $uri =404; }\n" {
		t.Errorf("nginx = %+v, want the store definition's", fw.Nginx)
	}
	if fw.Workers["pulse"].Command != "php acme pulse" {
		t.Error("the overlay's worker was not merged")
	}
}

// requires is honoured from a user overlay, as the docs promise, and replaces
// the base's list when set.
func TestGetFramework_OverlayRequiresReplacesTheBase(t *testing.T) {
	setConfigDir(t)
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	writeOverlay(t, "laravel", "name: laravel\nrequires: [opensearch]\n")

	fw, _ := GetFramework("laravel")
	if len(fw.Requires) != 1 || fw.Requires[0] != "opensearch" {
		t.Errorf("requires = %v, want the overlay's [opensearch]", fw.Requires)
	}
}

// lerd framework add saves a laravel overlay through SaveFramework, which must
// keep every key the merge honours or the overlay loses them before it is read.
func TestSaveFramework_LaravelOverlayKeepsNginxAndRequires(t *testing.T) {
	setConfigDir(t)
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	in := &Framework{
		Name:     "laravel",
		Nginx:    &FrameworkNginx{Snippet: "location = /vm-overlay { return 204; }\n"},
		Requires: []string{"opensearch"},
	}
	if err := SaveFramework(in); err != nil {
		t.Fatal(err)
	}

	fw, _ := GetFramework("laravel")
	if fw.Nginx == nil || fw.Nginx.Snippet != in.Nginx.Snippet {
		t.Errorf("nginx = %+v, want the saved overlay's", fw.Nginx)
	}
	if len(fw.Requires) != 1 || fw.Requires[0] != "opensearch" {
		t.Errorf("requires = %v, want [opensearch]", fw.Requires)
	}
}
