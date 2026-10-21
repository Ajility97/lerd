package plugins

import _ "embed"

// Guide is how to build a plugin, written for an AI assistant: the MCP plugin
// tool returns it and lerd mcp:inject writes it beside lerd's own skill.
//
//go:embed guide.md
var Guide string
