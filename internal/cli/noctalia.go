package cli

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/geodro/lerd/internal/config"
)

// noctaliaTemplate renders both of the palette's variants so lerd's light and
// dark modes each get theirs. Borders take a surface step, since community
// palettes leave outline_variant to the outline, far heavier than lerd's lines.
const noctaliaTemplate = `accent: "{{colors.primary.light.hex}}"
accent_dark: "{{colors.primary.dark.hex}}"
bg: "{{colors.surface.dark.hex}}"
card: "{{colors.surface_container.dark.hex}}"
border: "{{colors.surface_container_highest.dark.hex}}"
muted: "{{colors.on_surface_variant.dark.hex}}"
bg_light: "{{colors.surface_container.light.hex}}"
card_light: "{{colors.surface.light.hex}}"
border_light: "{{colors.surface_container_highest.light.hex}}"
`

// noctaliaRegistration is the user template entry. Noctalia expands the XDG
// variables itself, falling back to the same defaults lerd does.
const noctaliaRegistration = `# Written by lerd install, removed by lerd uninstall.
[theme.templates.user.lerd]
input_path = "$XDG_CONFIG_HOME/lerd/noctalia/lerd.yaml"
output_path = "$XDG_STATE_HOME/lerd/noctalia/theme.yaml"
`

// runNoctaliaMsg sends a command to the running shell, indirected for tests.
var runNoctaliaMsg = func(args ...string) error {
	return exec.Command("noctalia", append([]string{"msg"}, args...)...).Run()
}

// installNoctaliaTemplate registers the lerd template with Noctalia, reporting
// whether there is a Noctalia to register with. A shell that is not running
// renders it when it next starts, so a failed reload is not an error.
func installNoctaliaTemplate() (bool, error) {
	if _, err := exec.LookPath("noctalia"); err != nil {
		return false, nil
	}
	// lerd-ui watches this directory before the first render lands in it.
	if err := os.MkdirAll(filepath.Dir(config.NoctaliaThemePath()), 0o755); err != nil {
		return true, err
	}
	for path, data := range map[string]string{
		config.NoctaliaTemplatePath():     noctaliaTemplate,
		config.NoctaliaRegistrationPath(): noctaliaRegistration,
	} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return true, err
		}
		if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
			return true, err
		}
	}
	// The reload picks up a new registration, but only applying renders a
	// template an upgrade changed.
	_ = runNoctaliaMsg("config-reload")
	_ = runNoctaliaMsg("templates-apply")
	return true, nil
}

// removeNoctaliaTemplate undoes installNoctaliaTemplate, along with the file it
// rendered. It reports whether there was a registration to remove.
func removeNoctaliaTemplate() (bool, error) {
	if _, err := os.Stat(config.NoctaliaRegistrationPath()); err != nil {
		return false, nil
	}
	var errs []error
	for _, path := range []string{config.NoctaliaRegistrationPath(), config.NoctaliaTemplatePath(), config.NoctaliaThemePath()} {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			errs = append(errs, err)
		}
	}
	_ = runNoctaliaMsg("config-reload")
	return true, errors.Join(errs...)
}
