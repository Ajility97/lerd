//go:build !windows

package traymenu

import (
	"testing"

	"github.com/geodro/lerd/internal/config"
)

func TestLocalSettings_ReadsTheGlobalConfig(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	cfg, err := config.LoadGlobal()
	if err != nil {
		t.Fatal(err)
	}
	cfg.LAN.Exposed = true
	cfg.SetDumpsEnabled(true)
	cfg.SetNotificationsEnabled(false)
	cfg.SetHighContrastTrayIcon(true)
	if err := config.SaveGlobal(cfg); err != nil {
		t.Fatal(err)
	}

	s := LocalSettings()

	if !s.LANExposed || s.LANServicesExposed || !s.Dumps || s.Notifications || !s.HighContrastIcon {
		t.Errorf("settings do not match the config: %+v", s)
	}
}
