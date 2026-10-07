package plugins

import (
	"path/filepath"
	"testing"
)

func TestScaffold_RenamesAndStaysValid(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "demo")
	if err := Scaffold("demo", dest); err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	p, err := ReadPlugin(dest)
	if err != nil {
		t.Fatalf("a scaffolded plugin must validate: %v", err)
	}
	if p.Name != "demo" {
		t.Errorf("name = %q, want demo", p.Name)
	}
}

func TestScaffold_RefusesExistingDirAndBadName(t *testing.T) {
	if err := Scaffold("demo", t.TempDir()); err == nil {
		t.Error("scaffolding over an existing dir must fail")
	}
	if err := Scaffold("Demo!", filepath.Join(t.TempDir(), "x")); err == nil {
		t.Error("an invalid name must fail")
	}
}
