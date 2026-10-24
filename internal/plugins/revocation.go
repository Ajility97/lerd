package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/geodro/lerd/internal/config"
)

// Revocation is the catalog pulling a plugin for a security problem, with the
// reason every user who has it is shown. Commits are the approved commits it
// covers; with none it covers every install of the name.
type Revocation struct {
	Name    string   `json:"name"`
	Reason  string   `json:"reason"`
	At      string   `json:"at"`
	Commits []string `json:"commits,omitempty"`
}

// RevokedError is why a revoked plugin did not load.
type RevokedError struct{ Revocation }

func (e RevokedError) Error() string {
	return fmt.Sprintf("revoked from the plugin catalog on %s: %s", e.At, e.Reason)
}

// covers reports whether the revocation applies to an install of the plugin, by
// its name, which it cannot change, and the commit its install records; one with
// no record is covered only by a revocation of every commit.
func (r Revocation) covers(name string) bool {
	if r.Name != name {
		return false
	}
	if len(r.Commits) == 0 {
		return true
	}
	src, ok := InstalledSource(name)
	return ok && slices.Contains(r.Commits, src.Commit)
}

// revocations reads the revocations from the catalog lerd already holds, never
// fetching, since every load of the plugins asks.
func revocations() []Revocation {
	c, err := readCatalogCache()
	if err != nil {
		return nil
	}
	return c.Revoked
}

// revocationFor returns the revocation stopping the plugin, if any: one that
// covers it and that the user has not chosen to keep using it through.
func revocationFor(name string, revoked []Revocation, kept map[string]string) (Revocation, bool) {
	for _, r := range revoked {
		if !r.covers(name) {
			continue
		}
		if kept[name] == r.At {
			return Revocation{}, false
		}
		return r, true
	}
	return Revocation{}, false
}

// keptPath holds each revoked plugin the user chose to keep, with the date of
// the revocation they answered. Not lerd's config: Load runs while that config
// is still being read, and reading it from here would wait on itself.
func keptPath() string { return filepath.Join(config.DataDir(), "plugin-revocations-kept.json") }

func readKept() map[string]string {
	kept := map[string]string{}
	if data, err := os.ReadFile(keptPath()); err == nil {
		_ = json.Unmarshal(data, &kept)
	}
	return kept
}

func writeKept(kept map[string]string) error {
	data, err := json.Marshal(kept)
	if err != nil {
		return err
	}
	config.GuardRealWrite(keptPath())
	if err := os.MkdirAll(filepath.Dir(keptPath()), 0o755); err != nil {
		return err
	}
	return os.WriteFile(keptPath(), data, 0o644)
}

// forgetKept drops a removed plugin's answer, so a reinstall is stopped again.
func forgetKept(name string) error {
	kept := readKept()
	if _, ok := kept[name]; !ok {
		return nil
	}
	delete(kept, name)
	return writeKept(kept)
}

// Revoked lists the installed plugins a revocation is stopping now.
func Revoked() []Revocation {
	revoked := revocations()
	if len(revoked) == 0 {
		return nil
	}
	kept := readKept()
	var out []Revocation
	for _, r := range revoked {
		// Installed at all, broken or not: a revoked plugin is often both.
		if _, err := os.Stat(filepath.Join(Dir(), r.Name, ManifestFile)); err != nil {
			continue
		}
		if stop, ok := revocationFor(r.Name, revoked, kept); ok {
			out = append(out, stop)
		}
	}
	return out
}

// KeepRevoked lets a revoked plugin run again because the user chose to keep
// it. The choice is tied to this revocation, so a later one asks again.
func KeepRevoked(name string) error {
	r, ok := revocationFor(name, revocations(), nil)
	if !ok {
		return fmt.Errorf("plugin %s is not revoked", name)
	}
	kept := readKept()
	kept[name] = r.At
	return writeKept(kept)
}
