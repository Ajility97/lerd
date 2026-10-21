package serviceops

import (
	"fmt"

	"github.com/geodro/lerd/internal/config"
	"github.com/geodro/lerd/internal/plugins"
)

// RemovePlugin removes a plugin together with every service installed from its
// presets, so nothing it brought keeps running. The services' data is kept, as
// the plugin's own is.
func RemovePlugin(name string) error {
	if !plugins.ValidName(name) {
		return fmt.Errorf("invalid plugin name %q", name)
	}
	installed, _ := config.ListCustomServices()
	for _, svc := range installed {
		if svc.Plugin != name {
			continue
		}
		if err := RemoveService(svc.Name, RemoveOptions{}, nil); err != nil {
			return fmt.Errorf("removing service %s, which plugin %s ships: %w", svc.Name, name, err)
		}
	}
	return plugins.Remove(name)
}
