package ui

import (
	"net/http"
	"os"

	"github.com/geodro/lerd/internal/cli"
)

// Registration runs through these so tests never write into the user-scope
// config of the AI assistants on the machine running them.
var (
	mcpEnable     = cli.RunMCPEnableGlobal
	mcpDisable    = cli.RunMCPDisableGlobal
	mcpConfigured = func() bool {
		home, err := os.UserHomeDir()
		return err == nil && cli.MCPGlobalConfigured(home)
	}
)

// handleSettingsMCP registers lerd's MCP server with every supported AI
// assistant, or removes it, exactly as mcp:enable-global and mcp:disable-global do.
func handleSettingsMCP(w http.ResponseWriter, r *http.Request) {
	enabled, ok := readEnabledBody(w, r)
	if !ok {
		return
	}
	apply := mcpDisable
	if enabled {
		apply = mcpEnable
	}
	if err := apply(); err != nil {
		writeJSON(w, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, map[string]any{"ok": true, "mcp_global": enabled})
}
