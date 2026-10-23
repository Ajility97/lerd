package serviceops

import (
	"strings"

	"github.com/geodro/lerd/internal/config"
)

// Seams for ensureDefaultDatabase, so tests decide the URL and the create.
var (
	connectionURLFn  = serviceConnectionURL
	createDatabaseFn func(svc, name string) (bool, error)
	databaseExistsFn func(svc, name string) (bool, error)
)

// Assigned here rather than above: creating a database starts its service,
// which reaches back into this file through EnsureDependencyRunning.
func init() {
	createDatabaseFn = CreateDatabase
	databaseExistsFn = func(svc, name string) (bool, error) {
		return EntityExists(svc, EntityFor(svc, "databases"), name)
	}
}

// ensureDefaultDatabase creates the database a service's connection URL names,
// which an image only creates on a fresh data directory; without it a dependent
// pointed there crash-loops on boot.
func ensureDefaultDatabase(name string) error {
	db := connectionDatabase(connectionURLFn(name))
	if db == "" {
		return nil
	}
	if _, err := createDatabaseFn(name, db); err != nil {
		// Two dependents starting together both find it missing, and the
		// second create fails on a database that is there by then.
		if exists, _ := databaseExistsFn(name, db); exists {
			return nil
		}
		return err
	}
	return nil
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
