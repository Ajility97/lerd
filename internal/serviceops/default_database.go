package serviceops

import (
	"strings"

	"github.com/geodro/lerd/internal/config"
)

// Seams for ensureDefaultDatabase, so tests decide the URL and the create.
var (
	connectionURLFn  = serviceConnectionURL
	createDatabaseFn func(svc, name string) (bool, error)
)

// Assigned here rather than above: creating a database starts its service,
// which reaches back into this file through EnsureDependencyRunning.
func init() { createDatabaseFn = CreateDatabase }

// ensureDefaultDatabase creates the database a service's connection URL names.
// The image only creates it while initialising a fresh data directory, so one
// that predates it, or had it dropped, lacks it, and a dependent pointed there
// crash-loops on boot.
func ensureDefaultDatabase(name string) error {
	db := connectionDatabase(connectionURLFn(name))
	if db == "" {
		return nil
	}
	_, err := createDatabaseFn(name, db)
	return err
}

func serviceConnectionURL(name string) string {
	if IsBuiltin(name) {
		return config.DefaultPresetConnectionURL(name)
	}
	svc, err := config.LoadCustomService(name)
	if err != nil {
		return ""
	}
	return svc.ConnectionURL
}

// connectionDatabase returns the database a connection URL names in its path.
// Parsed by hand because a preset's URL can still carry {{host_port}}, which
// net/url refuses as a port.
func connectionDatabase(raw string) string {
	_, rest, ok := strings.Cut(raw, "://")
	if !ok {
		return ""
	}
	_, path, ok := strings.Cut(rest, "/")
	if !ok {
		return ""
	}
	path, _, _ = strings.Cut(path, "?")
	return strings.Trim(path, "/")
}
