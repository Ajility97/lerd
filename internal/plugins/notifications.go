package plugins

import (
	"fmt"

	"github.com/geodro/lerd/internal/config"
)

// The states of a plugin's notifications, as every surface shows them.
const (
	NotificationsNone    = "none"    // the plugin does not send any
	NotificationsPending = "pending" // it asks to, and the user has not answered
	NotificationsAllowed = "allowed"
	NotificationsDenied  = "denied"
)

// NotificationState is where the plugin stands. It is derived rather than
// stored, so an install or an update that starts asking is pending on its own,
// and an answer once given holds across updates.
func NotificationState(p *Plugin) string {
	if p == nil || !p.Notifications {
		return NotificationsNone
	}
	cfg, err := config.LoadGlobal()
	if err != nil {
		return NotificationsPending
	}
	allowed, asked := cfg.PluginNotifications(p.Name)
	switch {
	case !asked:
		return NotificationsPending
	case allowed:
		return NotificationsAllowed
	}
	return NotificationsDenied
}

// SetNotifications records the user's answer for an installed plugin that asks
// to send notifications.
func SetNotifications(name string, allowed bool) error {
	p, err := Find(name)
	if err != nil {
		return err
	}
	if !p.Notifications {
		return fmt.Errorf("plugin %s does not send notifications", name)
	}
	return config.UpdateGlobal(func(c *config.GlobalConfig) { c.SetPluginNotifications(name, allowed) })
}
