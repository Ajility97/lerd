package plugins

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

const revokedCommit = "0123456789abcdef0123456789abcdef01234567"

// installRevokable lays out an installed catalog plugin at revokedCommit and a
// cached catalog carrying the given revocations.
func installRevokable(t *testing.T, revoked ...Revocation) {
	t.Helper()
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	dir := writePlugin(t, Dir(), "bad", "api_version: 1\nname: bad\ncommands:\n  - {name: run, run: bin/run}\n")
	if err := writeSource(dir, &Source{Repo: "https://github.com/acme/bad", Commit: revokedCommit}); err != nil {
		t.Fatal(err)
	}
	setRevocations(t, revoked...)
}

func setRevocations(t *testing.T, revoked ...Revocation) {
	t.Helper()
	data, err := json.Marshal(Catalog{Revoked: revoked})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(catalogCachePath()), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(catalogCachePath(), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func loaded(name string) bool {
	list, _ := Load()
	for _, p := range list {
		if p.Name == name {
			return true
		}
	}
	return false
}

// A revoked plugin runs nothing: it does not load, it cannot be found to run
// one of its surfaces, and it says why.
func TestRevocation_StopsThePlugin(t *testing.T) {
	installRevokable(t, Revocation{Name: "bad", Reason: "sends .env away", At: "2026-10-24"})

	if loaded("bad") {
		t.Fatal("a revoked plugin loaded")
	}
	_, errs := Load()
	var rev RevokedError
	if len(errs) != 1 || !errors.As(errs[0].Err, &rev) || rev.Reason != "sends .env away" {
		t.Fatalf("load errors %v, want the revocation with its reason", errs)
	}
	if _, err := Find("bad"); !errors.As(err, &rev) {
		t.Fatalf("Find = %v, want the revocation", err)
	}
	if got := Revoked(); len(got) != 1 || got[0].Name != "bad" {
		t.Fatalf("Revoked = %v", got)
	}
}

// Keeping it holds through that revocation only; a later one asks again.
func TestRevocation_KeepHoldsUntilTheNextRevocation(t *testing.T) {
	installRevokable(t, Revocation{Name: "bad", Reason: "x", At: "2026-10-24"})
	if err := KeepRevoked("bad"); err != nil {
		t.Fatal(err)
	}
	if !loaded("bad") || len(Revoked()) != 0 {
		t.Fatal("a kept plugin must load again and leave the revoked list")
	}
	setRevocations(t, Revocation{Name: "bad", Reason: "worse", At: "2026-11-02"})
	if loaded("bad") {
		t.Fatal("a new revocation must stop it again")
	}
}

// Listing commits leaves an install at any other commit alone, which is how a
// fixed commit is approved again while the old ones stay revoked.
func TestRevocation_CoversOnlyTheCommitsItNames(t *testing.T) {
	installRevokable(t, Revocation{Name: "bad", Reason: "x", At: "2026-10-24", Commits: []string{"1111111111111111111111111111111111111111"}})
	if !loaded("bad") {
		t.Fatal("an install at a commit the revocation does not name must keep running")
	}
	setRevocations(t, Revocation{Name: "bad", Reason: "x", At: "2026-10-24", Commits: []string{revokedCommit}})
	if loaded("bad") {
		t.Fatal("an install at a named commit must stop")
	}
}

// Removing the plugin forgets that it was kept, so a reinstall starts clean.
func TestRevocation_RemoveForgetsTheKeep(t *testing.T) {
	installRevokable(t, Revocation{Name: "bad", Reason: "x", At: "2026-10-24"})
	if err := KeepRevoked("bad"); err != nil {
		t.Fatal(err)
	}
	if err := Remove("bad"); err != nil {
		t.Fatal(err)
	}
	dir := writePlugin(t, Dir(), "bad", "api_version: 1\nname: bad\n")
	if err := writeSource(dir, &Source{Repo: "https://github.com/acme/bad", Commit: revokedCommit}); err != nil {
		t.Fatal(err)
	}
	if loaded("bad") {
		t.Fatal("a reinstall must be stopped again")
	}
}
