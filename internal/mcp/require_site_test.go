package mcp

import (
	"fmt"
	"strings"
	"testing"

	"github.com/geodro/lerd/internal/config"
)

func TestRequireSite(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := config.AddSite(config.Site{Name: "shop", Domains: []string{"shop.test"}, Path: t.TempDir()}); err != nil {
		t.Fatal(err)
	}
	if _, _, errResp := requireSite(map[string]any{}); !strings.Contains(fmt.Sprint(errResp), "site is required") {
		t.Errorf("no site arg = %v", errResp)
	}
	if _, _, errResp := requireSite(map[string]any{"site": "nope"}); !strings.Contains(fmt.Sprint(errResp), "site not found: nope") {
		t.Errorf("unknown site = %v", errResp)
	}
	name, site, errResp := requireSite(map[string]any{"site": "shop"})
	if errResp != nil || name != "shop" || site == nil || site.Name != "shop" {
		t.Errorf("known site = %q %v %v", name, site, errResp)
	}
}
