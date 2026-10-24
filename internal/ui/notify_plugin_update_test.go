package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/geodro/lerd/internal/plugins"
)

// cacheCatalog serves listings as the registry's index and loads them into
// lerd's cache, the copy the dashboard reads updates from.
func cacheCatalog(t *testing.T, listings ...plugins.Listing) {
	t.Helper()
	serveCatalog(t, listings...)
	if err := plugins.RefreshCatalogOlderThan(0); err != nil {
		t.Fatal(err)
	}
}

// installedFromCatalog marks the echo plugin as installed from the catalog at commit.
func installedFromCatalog(t *testing.T, commit string) {
	t.Helper()
	data, _ := json.Marshal(map[string]string{"repo": "https://github.com/acme/echo", "commit": commit})
	if err := os.WriteFile(filepath.Join(plugins.Dir(), "echo", ".lerd-source.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestNewPluginUpdates_AnnouncesEachApprovedCommitOnce(t *testing.T) {
	installAPIPlugin(t)
	installedFromCatalog(t, strings.Repeat("a", 40))
	cacheCatalog(t, plugins.Listing{Name: "echo", Repo: "https://github.com/acme/echo", Commit: strings.Repeat("b", 40), Version: "0.2.0"})

	got := newPluginUpdates()
	if len(got) != 1 || got[0].name != "echo" || got[0].version != "0.2.0" {
		t.Fatalf("first check = %+v", got)
	}
	if got := newPluginUpdates(); len(got) != 0 {
		t.Errorf("an announced commit is not announced again, got %+v", got)
	}
	cacheCatalog(t, plugins.Listing{Name: "echo", Repo: "https://github.com/acme/echo", Commit: strings.Repeat("c", 40), Version: "0.3.0"})
	again := newPluginUpdates()
	if len(again) != 1 || again[0].version != "0.3.0" {
		t.Fatalf("a newer approved commit is news again, got %+v", again)
	}
	n := pluginUpdateNotification(again[0])
	if n.Kind != "update_available" || n.URL != "#plugins" || n.Params["version"] != "0.3.0" {
		t.Errorf("notification = %+v", n)
	}
}

func TestPluginManager_DescribesAnUpdate(t *testing.T) {
	installAPIPlugin(t)
	installedFromCatalog(t, strings.Repeat("a", 40))
	cacheCatalog(t, plugins.Listing{Name: "echo", Repo: "https://github.com/acme/echo", Commit: strings.Repeat("b", 40), Version: "0.2.0"})
	r := managedPlugins()[0]
	if r.Update != strings.Repeat("b", 40) || r.UpdateVersion != "0.2.0" ||
		r.Changes != "https://github.com/acme/echo/compare/"+strings.Repeat("a", 40)+"..."+strings.Repeat("b", 40) {
		t.Errorf("row = %+v", r)
	}
	rec := httptest.NewRecorder()
	handlePluginsList(rec, localRequest(http.MethodGet, "/api/plugins", ""))
	if !strings.Contains(rec.Body.String(), `"update":true`) {
		t.Errorf("the plugin list marks it for the rail's badge: %s", rec.Body)
	}
}

// A revoked commit does not load, but the fixed commit the catalog approves
// must still be offered, or the only way out is a remove and reinstall.
func TestPluginManager_OffersTheFixToARevokedPlugin(t *testing.T) {
	installAPIPlugin(t)
	installedFromCatalog(t, strings.Repeat("a", 40))
	catalog := plugins.Catalog{
		Plugins: []plugins.Listing{{Name: "echo", Repo: "https://github.com/acme/echo", Commit: strings.Repeat("b", 40), Version: "0.2.0"}},
		Revoked: []plugins.Revocation{{Name: "echo", Reason: "x", At: "2026-10-24", Commits: []string{strings.Repeat("a", 40)}}},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(catalog)
	}))
	t.Cleanup(srv.Close)
	orig := plugins.CatalogURL
	plugins.CatalogURL = srv.URL
	t.Cleanup(func() { plugins.CatalogURL = orig })
	if err := plugins.RefreshCatalogOlderThan(0); err != nil {
		t.Fatal(err)
	}

	r := managedPlugins()[0]
	if r.Revoked == nil || r.Update != strings.Repeat("b", 40) || r.UpdateVersion != "0.2.0" || r.Commit != strings.Repeat("a", 40) {
		t.Errorf("row = %+v", r)
	}
}
