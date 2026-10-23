package plugins

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDocsAndScreenshots_Validated(t *testing.T) {
	root := t.TempDir()
	dir := writePlugin(t, root, "p", "api_version: 1\nname: p\ndocs: DOCS.md\nscreenshots: [docs/tab.png]\n")
	writeFile(t, filepath.Join(dir, "docs", "tab.png"), "png")
	writeFile(t, filepath.Join(dir, "DOCS.md"), "# P")
	if _, err := ReadPlugin(dir); err != nil {
		t.Fatalf("docs and a screenshot inside the plugin must load: %v", err)
	}

	for name, c := range map[string]struct{ keys, want string }{
		"missing screenshot":   {"screenshots: [docs/nope.png]", "docs/nope.png"},
		"screenshot not image": {"screenshots: [bin/run]", "bin/run"},
		"screenshot outside":   {"screenshots: [../q/x.png]", "../q/x.png"},
		"missing docs":         {"docs: DOCS.md", "DOCS.md"},
		"docs not markdown":    {"docs: bin/run", "bin/run"},
		"icon not svg":         {"icon: bin/run", "bin/run"},
		"missing icon":         {"icon: icon.svg", "icon.svg"},
	} {
		t.Run(name, func(t *testing.T) {
			dir := writePlugin(t, t.TempDir(), "p", "api_version: 1\nname: p\n"+c.keys+"\n")
			_, err := ReadPlugin(dir)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("err = %v, want one naming %s", err, c.want)
			}
		})
	}
}

func TestServes_OnlyUIAndScreenshots(t *testing.T) {
	p := &Plugin{Manifest: Manifest{Icon: "icon.svg", Screenshots: []string{"docs/tab.png"}}}
	for rel, want := range map[string]bool{
		"icon.svg":        true,
		"ui/index.html":   true,
		"docs/tab.png":    true,
		"docs/other.png":  false,
		"plugin.yaml":     false,
		"bin/run":         false,
		"ui/../bin/run":   false,
		"docs/../tab.png": false,
	} {
		if got := p.Serves(rel); got != want {
			t.Errorf("Serves(%q) = %v, want %v", rel, got, want)
		}
	}
}

func TestFeatures_ListWhatTheManifestDeclares(t *testing.T) {
	p := &Plugin{Manifest: Manifest{
		Name:         "vd",
		Commands:     []Command{{Name: "check", Usage: "Compare with the baseline"}},
		SiteCommands: []SiteCommand{{Name: "check", Label: "Visual check", Description: "Screenshot this site"}},
		MCP:          &MCP{Tools: []MCPTool{{Name: "check", Description: "Check a site"}}},
		Hooks:        map[string]string{"site.linked": "bin/run"},
		UI:           &UI{SiteTab: &UIPage{Title: "Visual diff"}, PageScript: "ui/page.js"},
	}}
	var got []string
	for _, f := range p.Features() {
		got = append(got, f.Kind+"|"+f.Name+"|"+f.Detail)
	}
	want := []string{
		"command|lerd vd:check|Compare with the baseline",
		"site_command|Visual check|Screenshot this site",
		"mcp|vd_check|Check a site",
		"hook|site.linked|",
		"site_tab|Visual diff|",
		"page_script||",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("Features =\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestDocs_RenderedForTheDashboard(t *testing.T) {
	dir := writePlugin(t, t.TempDir(), "p", "api_version: 1\nname: p\ndocs: DOCS.md\nscreenshots: [docs/tab.png]\n")
	writeFile(t, filepath.Join(dir, "docs", "tab.png"), "png")
	writeFile(t, filepath.Join(dir, "DOCS.md"), strings.Join([]string{
		"# P",
		"![tab](docs/tab.png) ![private](bin/secret.png)",
		"See [the guide](docs/guide.md), [above](#p) and [lerd](https://lerd.sh).",
		"<script>alert(1)</script>",
	}, "\n\n"))
	p, err := ReadPlugin(dir)
	if err != nil {
		t.Fatal(err)
	}
	html, err := p.DocsHTML(func(rel string) string { return "/plugins/p/" + rel })
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`<img src="/plugins/p/docs/tab.png" alt="tab"`, `href="https://lerd.sh" target="_blank" rel="noopener noreferrer"`, "the guide", "above"} {
		if !strings.Contains(html, want) {
			t.Errorf("docs should contain %q:\n%s", want, html)
		}
	}
	// A file lerd does not serve, a link into the repository and an anchor that
	// would move the dashboard's own route all lose their target.
	for _, bad := range []string{"bin/secret.png", "docs/guide.md", `href="#p"`, "<script>"} {
		if strings.Contains(html, bad) {
			t.Errorf("docs should not contain %q:\n%s", bad, html)
		}
	}
}

func TestDocs_EmptyWhenNoneDeclared(t *testing.T) {
	p, err := ReadPlugin(writePlugin(t, t.TempDir(), "p", "api_version: 1\nname: p\n"))
	if err != nil {
		t.Fatal(err)
	}
	if html, err := p.DocsHTML(func(string) string { return "" }); err != nil || html != "" {
		t.Errorf("Docs = %q, %v; want empty and no error", html, err)
	}
}

func TestCardIcon_OwnIconBeforeTheRailPages(t *testing.T) {
	page := &UI{Page: &UIPage{Icon: "ui/icon.svg"}}
	for _, c := range []struct {
		m    Manifest
		want string
	}{
		{Manifest{Icon: "icon.svg", UI: page}, "icon.svg"},
		{Manifest{UI: page}, "ui/icon.svg"},
		{Manifest{}, ""},
	} {
		if got := (&Plugin{Manifest: c.m}).CardIcon(); got != c.want {
			t.Errorf("CardIcon = %q, want %q", got, c.want)
		}
	}
}

// An image is resolved from the docs file's own folder, the way the markdown
// reads where it sits.
func TestDocs_ResolvesImagesFromTheDocsFolder(t *testing.T) {
	dir := writePlugin(t, t.TempDir(), "p", "api_version: 1\nname: p\ndocs: docs/README.md\nscreenshots: [docs/tab.png]\n")
	writeFile(t, filepath.Join(dir, "docs", "tab.png"), "png")
	writeFile(t, filepath.Join(dir, "docs", "README.md"), "![tab](tab.png)")
	p, err := ReadPlugin(dir)
	if err != nil {
		t.Fatal(err)
	}
	html, err := p.DocsHTML(func(rel string) string { return "/plugins/p/" + rel })
	if err != nil || !strings.Contains(html, `src="/plugins/p/docs/tab.png"`) {
		t.Errorf("html = %s, %v", html, err)
	}
}
