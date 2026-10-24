package cli

import (
	"errors"
	"os/exec"
	"strings"

	"github.com/geodro/lerd/internal/feedback"
	"github.com/geodro/lerd/internal/platform"
)

// errNoClaude is a Windows with no claude CLI, which is no reason to fail.
var errNoClaude = errors.New("claude is not installed")

// windowsAIHome is the Windows profile, whose AI clients run lerd.exe by its
// full path, so one already open needs no fresh PATH to find it.
func windowsAIHome() (aiHome, error) {
	profile, err := windowsEnvPath("USERPROFILE")
	if err != nil {
		return aiHome{}, err
	}
	// A Windows path, not wslpath's answer: uninstall takes the registration
	// back after lerd.exe is gone, and wslpath cannot name a missing file.
	local, err := powershell("$env:LOCALAPPDATA")
	if err != nil {
		return aiHome{}, err
	}
	exe := strings.TrimSpace(local) + `\lerd\bin\lerd.exe`
	return aiHome{dir: profile, command: exe, claude: windowsClaudeMCP, claudeDir: windowsClaudeConfigDir(), windows: true}, nil
}

// windowsClaudeConfigDir is the Windows user's CLAUDE_CONFIG_DIR as a WSL path,
// "" when unset. It is read from the registry, since the environment WSL hands
// PowerShell is the one Windows had when the distro started.
func windowsClaudeConfigDir() string {
	out, err := powershell("[Environment]::GetEnvironmentVariable('CLAUDE_CONFIG_DIR', 'User')")
	if win := strings.TrimSpace(out); err == nil && win != "" {
		if p, err := exec.Command("wslpath", "-u", win).Output(); err == nil {
			return strings.TrimSpace(string(p))
		}
	}
	return ""
}

// windowsClaudeMCP runs the Windows claude CLI's mcp subcommand, an exe or the
// claude.cmd npm installs, whichever PowerShell finds.
func windowsClaudeMCP(args ...string) ([]byte, error) {
	out, err := powershell(windowsClaudeScript(args))
	if err == nil && strings.TrimSpace(out) == "no-claude" {
		return nil, errNoClaude
	}
	return []byte(out), err
}

func windowsClaudeScript(args []string) string {
	quoted := make([]string, len(args))
	for i, a := range args {
		quoted[i] = psQuote(a)
	}
	// WSL hands PowerShell the PATH Windows had when the distro started, which
	// misses a claude installed since, and Claude's own installer leaves its
	// folder off PATH for the user to add, so that folder is looked in too.
	return "$env:Path = [Environment]::GetEnvironmentVariable('Path', 'User') + ';' + [Environment]::GetEnvironmentVariable('Path', 'Machine')\n" +
		"$c = (Get-Command claude -CommandType Application -ErrorAction SilentlyContinue | Select-Object -First 1).Source\n" +
		"if (-not $c) { $c = Join-Path $env:USERPROFILE '.local\\bin\\claude.exe'; if (-not (Test-Path $c)) { 'no-claude'; exit 0 } }\n" +
		"& $c mcp " + strings.Join(quoted, " ") + "\nexit $LASTEXITCODE"
}

// withWindowsAIHome runs fn on the Windows profile of a WSL install, warning
// rather than failing when Windows cannot be reached, as the Linux side is done.
func withWindowsAIHome(fn func(aiHome) error) {
	if !platform.Current.WindowsSide {
		return
	}
	h, err := windowsAIHome()
	if err == nil {
		err = fn(h)
	}
	if err != nil {
		feedback.Warn("Windows AI clients: %v", err)
	}
}

// enableWindowsAIClients registers lerd with the AI clients installed on Windows.
func enableWindowsAIClients(verbose bool) {
	withWindowsAIHome(func(h aiHome) error {
		if err := writeHomeMCPConfigs(h, verbose); err != nil {
			return err
		}
		return writeHomeContexts(h, verbose, true)
	})
}

// refreshWindowsAIClients keeps the Windows clients' lerd docs in step with
// this lerd, writing only the ones already there.
func refreshWindowsAIClients(verbose bool) {
	withWindowsAIHome(func(h aiHome) error { return writeHomeContexts(h, verbose, false) })
}

func removeWindowsAIClients(verbose bool) {
	withWindowsAIHome(func(h aiHome) error { return removeHomeAI(h, verbose) })
}
