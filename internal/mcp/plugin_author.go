package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/geodro/lerd/internal/plugins"
)

// pluginAuthorTool lets an assistant build a plugin for the user. Installing
// from a git URL or the catalog stays with the user: only a local folder.
func pluginAuthorTool() mcpTool {
	return mcpTool{
		Name:        "plugin",
		Description: `Build a lerd plugin for the user; read guide first. new scaffolds one, list shows load errors, install copies a local folder.`,
		InputSchema: mcpSchema{
			Type: "object",
			Properties: map[string]mcpProp{
				"action": {Type: "string", Enum: []string{"guide", "new", "list", "install"}},
				"name":   {Type: "string", Description: "new: plugin name."},
				"path":   {Type: "string", Description: "install: absolute folder path."},
			},
			Required: []string{"action"},
		},
	}
}

func execPluginGuide(map[string]any) (any, *rpcError) {
	return toolOK(plugins.Guide), nil
}

func execPluginNew(args map[string]any) (any, *rpcError) {
	name := strArg(args, "name")
	dest := filepath.Join(plugins.Dir(), name)
	if err := plugins.Scaffold(name, dest); err != nil {
		return toolErr(err.Error()), nil
	}
	return toolOK("Created plugin " + name + " in " + dest + ". Edit it there; changes apply at once. Read the plugin guide before editing."), nil
}

func execPluginInstall(args map[string]any) (any, *rpcError) {
	path := strArg(args, "path")
	if info, err := os.Stat(path); err != nil || !info.IsDir() || !filepath.IsAbs(path) {
		return toolErr("path must be the absolute path of a plugin folder: " + path), nil
	}
	p, err := plugins.Install(path)
	if err != nil {
		return toolErr(err.Error()), nil
	}
	return toolOK("Installed plugin " + p.Name + " in " + p.Dir), nil
}

func execPluginList(map[string]any) (any, *rpcError) {
	type entry struct {
		Name     string   `json:"name"`
		Dir      string   `json:"dir"`
		Surfaces []string `json:"surfaces"`
	}
	type broken struct {
		Dir   string `json:"dir"`
		Error string `json:"error"`
	}
	list, errs := plugins.Load()
	out := struct {
		Installed []entry  `json:"installed"`
		Broken    []broken `json:"broken"`
	}{Installed: []entry{}, Broken: []broken{}}
	for _, p := range list {
		out.Installed = append(out.Installed, entry{Name: p.Name, Dir: p.Dir, Surfaces: surfacesOf(p)})
	}
	for _, e := range errs {
		out.Broken = append(out.Broken, broken{Dir: e.Dir, Error: e.Err.Error()})
	}
	data, _ := json.MarshalIndent(out, "", "  ")
	return toolOK(string(data)), nil
}

// surfacesOf names what a loaded plugin adds, so an assistant can check its
// edits landed without reading the manifest back.
func surfacesOf(p *plugins.Plugin) []string {
	var s []string
	add := func(on bool, name string) {
		if on {
			s = append(s, name)
		}
	}
	add(len(p.Commands) > 0, "commands")
	add(len(p.SiteCommands) > 0, "site_commands")
	add(len(p.Hooks) > 0, "hooks")
	add(p.MCP != nil && len(p.MCP.Tools) > 0, "mcp")
	add(len(p.API) > 0, "api")
	if p.UI != nil {
		add(p.UI.Page != nil, "page")
		add(p.UI.SiteTab != nil, "site_tab")
		add(p.UI.OverviewCard != nil, "overview_card")
		add(p.UI.PageScript != "", "page_script")
	}
	return s
}
