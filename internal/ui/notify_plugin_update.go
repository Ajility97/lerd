package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/geodro/lerd/internal/config"
	"github.com/geodro/lerd/internal/plugins"
	"github.com/geodro/lerd/internal/push"
)

// pluginUpdateEvery is how often lerd-ui re-reads the catalog for newer
// approved commits of the plugins installed from it.
const pluginUpdateEvery = time.Hour

// watchPluginUpdates refreshes the catalog every hour and announces each
// newly approved commit of an installed plugin once.
func watchPluginUpdates() {
	go func() {
		// Start-up has more pressing work than the catalog.
		time.Sleep(time.Minute)
		for {
			_ = plugins.RefreshCatalogOlderThan(pluginUpdateEvery)
			broker.broadcastPlugins()
			for _, u := range newPluginUpdates() {
				dispatchNotification(pluginUpdateNotification(u))
			}
			for _, r := range newRevocations() {
				dispatchNotification(revocationNotification(r))
			}
			time.Sleep(pluginUpdateEvery)
		}
	}()
}

type pluginUpdate struct {
	name, version, commit string
}

func announcedUpdatesPath() string {
	return filepath.Join(config.DataDir(), "plugin-updates-announced.json")
}

// newPluginUpdates returns the updates not announced before and remembers
// them, so each approved commit is announced once, across restarts too.
func newPluginUpdates() []pluginUpdate {
	list, _ := plugins.Load()
	announced := map[string]string{}
	if data, err := os.ReadFile(announcedUpdatesPath()); err == nil {
		_ = json.Unmarshal(data, &announced)
	}
	var out []pluginUpdate
	for _, p := range list {
		l, ok := plugins.CachedUpdate(p.Name)
		if !ok || announced[p.Name] == l.Commit {
			continue
		}
		announced[p.Name] = l.Commit
		out = append(out, pluginUpdate{name: p.Name, version: l.Version, commit: l.Commit})
	}
	if len(out) > 0 {
		if data, err := json.Marshal(announced); err == nil {
			path := announcedUpdatesPath()
			config.GuardRealWrite(path)
			_ = os.WriteFile(path, data, 0o644)
		}
	}
	return out
}

func pluginUpdateNotification(u pluginUpdate) push.Notification {
	version := u.version
	if version == "" {
		version = u.commit[:7]
	}
	return push.Notification{
		Kind:     "update_available",
		TitleKey: "notify_update_title",
		Title:    "Update available: " + u.name,
		BodyKey:  "notify_plugin_update_body",
		Body:     "Version " + version + " is approved. Update it from the Plugins page.",
		Params:   map[string]string{"service": u.name, "version": version},
		Tag:      "lerd-update-plugin-" + u.name,
		URL:      "#plugins",
		Data:     map[string]string{"plugin": u.name, "version": version},
		Urgency:  "low",
		TTL:      3600,
	}
}

func announcedRevocationsPath() string {
	return filepath.Join(config.DataDir(), "plugin-revocations-announced.json")
}

// newRevocations returns the revocations stopping an installed plugin that
// were not announced before, so each one is announced once. The dashboard's
// banner is what holds the user to it; this only makes sure they hear.
func newRevocations() []plugins.Revocation {
	announced := map[string]string{}
	if data, err := os.ReadFile(announcedRevocationsPath()); err == nil {
		_ = json.Unmarshal(data, &announced)
	}
	var out []plugins.Revocation
	for _, r := range plugins.Revoked() {
		if announced[r.Name] == r.At {
			continue
		}
		announced[r.Name] = r.At
		out = append(out, r)
	}
	if len(out) > 0 {
		if data, err := json.Marshal(announced); err == nil {
			path := announcedRevocationsPath()
			config.GuardRealWrite(path)
			_ = os.WriteFile(path, data, 0o644)
		}
	}
	return out
}

func revocationNotification(r plugins.Revocation) push.Notification {
	return push.Notification{
		Kind:     "plugin_failed",
		TitleKey: "notify_plugin_revoked_title",
		Title:    "Plugin stopped: " + r.Name,
		BodyKey:  "notify_plugin_revoked_body",
		Body:     r.Reason,
		Params:   map[string]string{"plugin": r.Name, "reason": r.Reason},
		Tag:      "lerd-plugin-revoked-" + r.Name,
		URL:      "#plugins",
		Data:     map[string]string{"plugin": r.Name},
		Urgency:  "high",
		TTL:      86400,
	}
}
