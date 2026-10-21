package siteinfo

import (
	"slices"
	"testing"

	"github.com/geodro/lerd/internal/config"
)

// The cards a site switched on travel with it, so a switch made from one
// dashboard reaches an Overview another one already has open.
func TestEnrich_CarriesTheOverviewCards(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("XDG_DATA_HOME", tmp)
	e := Enrich(config.Site{Name: "shop", Domains: []string{"shop.test"}, Path: t.TempDir(), OverviewCards: []string{"stats"}}, 0)
	if !slices.Equal(e.OverviewCards, []string{"stats"}) {
		t.Errorf("OverviewCards = %v, want [stats]", e.OverviewCards)
	}
}
