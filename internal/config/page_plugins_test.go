package config

import (
	"slices"
	"testing"
)

// The switch survives a round trip through sites.yaml, switching one on twice
// lists it once, and switching the last one off leaves nothing behind.
func TestSetSitePagePlugin_TogglesAndPersists(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("XDG_DATA_HOME", tmp)
	if err := AddSite(Site{Name: "shop", Domains: []string{"shop.test"}, Path: t.TempDir()}); err != nil {
		t.Fatal(err)
	}

	steps := []struct {
		plugin string
		on     bool
		want   []string
	}{
		{"debug", true, []string{"debug"}},
		{"debug", true, []string{"debug"}},
		{"notes", true, []string{"debug", "notes"}},
		{"debug", false, []string{"notes"}},
		{"notes", false, nil},
	}
	for _, s := range steps {
		if err := SetSitePagePlugin("shop", s.plugin, s.on); err != nil {
			t.Fatal(err)
		}
		invalidateSitesCache()
		site, err := FindSite("shop")
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(site.PagePlugins, s.want) {
			t.Fatalf("after %s=%v: PagePlugins = %v, want %v", s.plugin, s.on, site.PagePlugins, s.want)
		}
	}
	if err := SetSitePagePlugin("ghost", "debug", true); err == nil {
		t.Fatal("expected an error for a site that is not registered")
	}
}

// Overview cards are switched apart from page scripts: one plugin can have its
// card on a site's Overview while its script stays off that site's pages.
func TestSetSiteOverviewCard_KeptApartFromPageScripts(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("XDG_DATA_HOME", tmp)
	if err := AddSite(Site{Name: "shop", Domains: []string{"shop.test"}, Path: t.TempDir()}); err != nil {
		t.Fatal(err)
	}
	if err := SetSiteOverviewCard("shop", "stats", true); err != nil {
		t.Fatal(err)
	}
	invalidateSitesCache()
	site, _ := FindSite("shop")
	if !slices.Equal(site.OverviewCards, []string{"stats"}) || site.PagePlugins != nil {
		t.Fatalf("OverviewCards = %v, PagePlugins = %v", site.OverviewCards, site.PagePlugins)
	}
	if err := SetSiteOverviewCard("shop", "stats", false); err != nil {
		t.Fatal(err)
	}
	invalidateSitesCache()
	site, _ = FindSite("shop")
	if site.OverviewCards != nil {
		t.Fatalf("OverviewCards = %v, want none", site.OverviewCards)
	}
}
