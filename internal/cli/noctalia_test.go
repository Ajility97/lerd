package cli

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/geodro/lerd/internal/config"
)

// withNoctalia puts a stub noctalia on PATH, or none, under a fresh config and
// state home, and records the messages sent to it.
func withNoctalia(t *testing.T, installed bool) *[][]string {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	bin := t.TempDir()
	t.Setenv("PATH", bin)
	if installed {
		if err := os.WriteFile(filepath.Join(bin, "noctalia"), []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	var calls [][]string
	prev := runNoctaliaMsg
	runNoctaliaMsg = func(args ...string) error {
		calls = append(calls, args)
		return nil
	}
	t.Cleanup(func() { runNoctaliaMsg = prev })
	return &calls
}

func TestInstallNoctaliaTemplate_RegistersTheTemplate(t *testing.T) {
	calls := withNoctalia(t, true)

	if found, err := installNoctaliaTemplate(); !found || err != nil {
		t.Fatalf("installNoctaliaTemplate = %v, %v, want an install", found, err)
	}
	reg, err := os.ReadFile(config.NoctaliaRegistrationPath())
	if err != nil {
		t.Fatal(err)
	}
	// Noctalia expands the variables itself, so the paths have to name the
	// same files lerd reads and writes.
	expand := func(s string) string {
		return strings.NewReplacer("$XDG_CONFIG_HOME", os.Getenv("XDG_CONFIG_HOME"), "$XDG_STATE_HOME", os.Getenv("XDG_STATE_HOME")).Replace(s)
	}
	for _, want := range []string{
		`input_path = "` + config.NoctaliaTemplatePath() + `"`,
		`output_path = "` + config.NoctaliaThemePath() + `"`,
	} {
		if !strings.Contains(expand(string(reg)), want) {
			t.Errorf("registration %q lacks %q", reg, want)
		}
	}
	if _, err := os.Stat(config.NoctaliaTemplatePath()); err != nil {
		t.Errorf("template not written: %v", err)
	}
	if want := [][]string{{"config-reload"}, {"templates-apply"}}; !reflect.DeepEqual(*calls, want) {
		t.Errorf("calls = %v, want %v", *calls, want)
	}
}

func TestInstallNoctaliaTemplate_NothingWithoutNoctalia(t *testing.T) {
	calls := withNoctalia(t, false)

	if found, _ := installNoctaliaTemplate(); found {
		t.Error("installNoctaliaTemplate claimed an install with no noctalia on PATH")
	}
	if _, err := os.Stat(config.NoctaliaRegistrationPath()); !os.IsNotExist(err) {
		t.Errorf("registration written without noctalia: %v", err)
	}
	if len(*calls) != 0 {
		t.Errorf("calls = %v, want none", *calls)
	}
}

func TestRemoveNoctaliaTemplate_RemovesWhatInstallWrote(t *testing.T) {
	calls := withNoctalia(t, true)
	if _, err := installNoctaliaTemplate(); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(config.NoctaliaThemePath()), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config.NoctaliaThemePath(), []byte("accent: \"#fb5bb6\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	*calls = nil

	if found, err := removeNoctaliaTemplate(); !found || err != nil {
		t.Fatalf("removeNoctaliaTemplate = %v, %v, want a removal", found, err)
	}
	for _, p := range []string{config.NoctaliaRegistrationPath(), config.NoctaliaTemplatePath(), config.NoctaliaThemePath()} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("%s still there: %v", p, err)
		}
	}
	if want := [][]string{{"config-reload"}}; !reflect.DeepEqual(*calls, want) {
		t.Errorf("calls = %v, want %v", *calls, want)
	}
}

func TestRemoveNoctaliaTemplate_NothingWhenNeverInstalled(t *testing.T) {
	calls := withNoctalia(t, true)

	if found, _ := removeNoctaliaTemplate(); found {
		t.Error("removeNoctaliaTemplate claimed a removal with nothing registered")
	}
	if len(*calls) != 0 {
		t.Errorf("calls = %v, want none", *calls)
	}
}
