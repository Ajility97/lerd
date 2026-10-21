package winshim

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

// Only the shell scripts lerd writes get a twin; the binaries and version
// stamps beside them in the bin dir do not.
func TestShimNames(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"php":                   "#!/bin/sh\nexec lerd php \"$@\"\n",
		"composer":              "#!/bin/sh\nexec lerd composer \"$@\"\n",
		"mkcert":                "\x7fELF binary",
		"composer.phar":         "<?php",
		"composer.phar.version": "2.8",
	})
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := ShimNames(dir)
	if err != nil || strings.Join(got, ",") != "composer,php" {
		t.Errorf("ShimNames = %q, %v; want [composer php]", got, err)
	}
}

// A shim dir that cannot be read is an error, not an empty list, or the agent
// would drop every twin while the distro is unreachable.
func TestShimNames_UnreadableDirFails(t *testing.T) {
	if _, err := ShimNames(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Error("a missing shim dir must fail")
	}
}

// php.exe and its siblings are lerd.exe under another name: a .cmd would lose
// the current folder, since cmd.exe refuses a UNC one like \\wsl.localhost.
func TestSyncTwins(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{"lerd.exe": "LERD", "node.exe": "LERD", "psql.exe": "OLD", "php.cmd": "stale", "lerd-wsl.conf": "distro=U"})
	if err := SyncTwins(dir, []string{"php", "composer", "psql"}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"php.exe", "composer.exe", "psql.exe", "lerd.exe"} {
		if b, err := os.ReadFile(filepath.Join(dir, name)); err != nil || string(b) != "LERD" {
			t.Errorf("%s = %q, %v; want lerd.exe's content", name, b, err)
		}
	}
	for _, name := range []string{"node.exe", "php.cmd"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			t.Errorf("%s left behind", name)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "lerd-wsl.conf")); err != nil {
		t.Error("the config must survive")
	}
}

// The agent syncs every half minute, so a twin already current stays the very
// file it was rather than being rewritten each time.
func TestSyncTwins_KeepsACurrentTwin(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{"lerd.exe": "LERD"})
	if err := SyncTwins(dir, []string{"php"}); err != nil {
		t.Fatal(err)
	}
	before, _ := os.Stat(filepath.Join(dir, "php.exe"))
	if err := SyncTwins(dir, []string{"php"}); err != nil {
		t.Fatal(err)
	}
	after, _ := os.Stat(filepath.Join(dir, "php.exe"))
	if !os.SameFile(before, after) {
		t.Error("a current twin was replaced")
	}
}
