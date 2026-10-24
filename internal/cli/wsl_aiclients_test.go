package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	toml "github.com/pelletier/go-toml/v2"
)

const testLerdExe = `C:\Users\dev\AppData\Local\lerd\bin\lerd.exe`

// windowsTestHome is a Windows profile where only the named client folders exist.
func windowsTestHome(t *testing.T, claude func(args ...string) ([]byte, error), dirs ...string) aiHome {
	t.Helper()
	dir := t.TempDir()
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return aiHome{dir: dir, command: testLerdExe, claude: claude, windows: true}
}

func readMCPEntry(t *testing.T, path, key string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	servers, _ := cfg[key].(map[string]any)
	entry, _ := servers["lerd"].(map[string]any)
	if entry == nil {
		t.Fatalf("%s has no lerd entry: %s", path, data)
	}
	return entry
}

func TestWindowsHome_registersOnlyInstalledClientsWithLerdExe(t *testing.T) {
	h := windowsTestHome(t, func(...string) ([]byte, error) { return nil, errNoClaude },
		".cursor", filepath.Join("AppData", "Roaming", "Code"))
	if err := writeHomeMCPConfigs(h, false); err != nil {
		t.Fatal(err)
	}

	cursor := readMCPEntry(t, filepath.Join(h.dir, ".cursor", "mcp.json"), "mcpServers")
	if cursor["command"] != testLerdExe {
		t.Errorf("cursor command = %v, want lerd.exe's Windows path", cursor["command"])
	}
	code := readMCPEntry(t, filepath.Join(h.dir, "AppData", "Roaming", "Code", "User", "mcp.json"), "servers")
	if code["command"] != testLerdExe || code["type"] != "stdio" {
		t.Errorf("VS Code entry = %v", code)
	}
	for _, rel := range []string{".codex", ".gemini", ".junie", ".codeium", ".config", ".claude"} {
		if _, err := os.Stat(filepath.Join(h.dir, rel)); !os.IsNotExist(err) {
			t.Errorf("%s was created for a client that is not installed", rel)
		}
	}
}

func TestWindowsHome_codexEntryKeepsTheWindowsPath(t *testing.T) {
	h := windowsTestHome(t, func(...string) ([]byte, error) { return nil, errNoClaude }, ".codex")
	if err := writeHomeMCPConfigs(h, false); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(h.dir, ".codex", "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		MCPServers map[string]struct {
			Command string   `toml:"command"`
			Args    []string `toml:"args"`
		} `toml:"mcp_servers"`
	}
	if err := toml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("config.toml does not parse: %v\n%s", err, data)
	}
	if got := cfg.MCPServers["lerd"].Command; got != testLerdExe {
		t.Errorf("codex command = %q, want %q", got, testLerdExe)
	}
}

func TestWindowsHome_claudeRegistersThroughItsCLI(t *testing.T) {
	var calls []string
	h := windowsTestHome(t, func(args ...string) ([]byte, error) {
		calls = append(calls, strings.Join(args, " "))
		return nil, nil
	}, ".claude")
	if err := writeHomeMCPConfigs(h, false); err != nil {
		t.Fatal(err)
	}
	want := "add --scope user lerd -- " + testLerdExe + " mcp"
	if len(calls) == 0 || calls[len(calls)-1] != want {
		t.Errorf("claude calls = %q, want the last to be %q", calls, want)
	}
}

func TestWindowsHome_noClaudeCLIIsNotAnError(t *testing.T) {
	h := windowsTestHome(t, func(...string) ([]byte, error) { return nil, errNoClaude }, ".claude")
	if err := writeHomeMCPConfigs(h, false); err != nil {
		t.Fatalf("a Windows without the claude CLI should be skipped, got %v", err)
	}
}

func TestWindowsHome_removeTakesBackEntriesAndSkills(t *testing.T) {
	var removed bool
	h := windowsTestHome(t, func(args ...string) ([]byte, error) {
		if args[0] == "remove" {
			removed = true
		}
		return nil, nil
	}, ".claude", ".cursor")
	if err := writeHomeMCPConfigs(h, false); err != nil {
		t.Fatal(err)
	}
	if err := writeHomeContexts(h, false, true); err != nil {
		t.Fatal(err)
	}
	skill := filepath.Join(h.dir, ".claude", "skills", "lerd", "SKILL.md")
	if _, err := os.Stat(skill); err != nil {
		t.Fatalf("the Claude skill was not written to the Windows profile: %v", err)
	}
	removed = false
	if err := removeHomeAI(h, false); err != nil {
		t.Fatal(err)
	}
	if !removed {
		t.Error("the Claude Code registration was not removed")
	}
	if _, err := os.Stat(skill); !os.IsNotExist(err) {
		t.Errorf("the skill is still there: %v", err)
	}
	if _, err := os.Stat(filepath.Join(h.dir, ".cursor", "mcp.json")); !os.IsNotExist(err) {
		t.Errorf("cursor's lerd-only config is still there: %v", err)
	}
}

func TestWindowsClaudeScript_quotesThePath(t *testing.T) {
	got := windowsClaudeScript([]string{"add", "--scope", "user", "lerd", "--", `C:\Users\O'Brien\lerd.exe`, "mcp"})
	if !strings.Contains(got, `'C:\Users\O''Brien\lerd.exe'`) {
		t.Errorf("path not quoted for PowerShell: %s", got)
	}
}

func TestWindowsHome_claudeConfigDirIsWhereTheSkillGoes(t *testing.T) {
	h := windowsTestHome(t, func(...string) ([]byte, error) { return nil, nil })
	h.claudeDir = filepath.Join(t.TempDir(), "claude-work")
	if err := os.MkdirAll(h.claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeHomeContexts(h, false, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(h.claudeDir, "skills", "lerd", "SKILL.md")); err != nil {
		t.Errorf("skill not under CLAUDE_CONFIG_DIR: %v", err)
	}
	if _, err := os.Stat(filepath.Join(h.dir, ".claude")); !os.IsNotExist(err) {
		t.Errorf("a default .claude folder was created beside the custom one")
	}
}

func TestWindowsClaudeScript_skipsPowerShellScripts(t *testing.T) {
	got := windowsClaudeScript([]string{"get", "lerd"})
	if !strings.Contains(got, "-CommandType Application") {
		t.Errorf("claude.ps1 could be picked, which PowerShell may refuse to run: %s", got)
	}
}
