package cli

import (
	"testing"

	"github.com/geodro/lerd/internal/config"
)

// The endpoint an app signs against and the URL the browser opens have to be
// the same host, so both shapes lerd writes have to land on the domain.
func TestApplyServiceDomainEnv_RepointsBothShapesAtTheDomain(t *testing.T) {
	updates := map[string]string{
		"AWS_ENDPOINT": "http://lerd-rustfs:9000",
		"AWS_URL":      "http://localhost:9000/uploads",
		"REDIS_HOST":   "lerd-redis",
		"MAIL_HOST":    "lerd-mailpit",
	}
	applyServiceDomainEnv(updates,
		map[string]string{"rustfs": "rustfs.test"},
		map[string]int{"rustfs": 9000}, nil)

	if got := updates["AWS_ENDPOINT"]; got != "https://rustfs.test" {
		t.Errorf("AWS_ENDPOINT = %q, want https://rustfs.test", got)
	}
	if got := updates["AWS_URL"]; got != "https://rustfs.test/uploads" {
		t.Errorf("AWS_URL = %q, want https://rustfs.test/uploads", got)
	}
	// A bare connection host is not a URL a browser ever sees, and rewriting it
	// would point the app at nginx instead of the service.
	if updates["REDIS_HOST"] != "lerd-redis" || updates["MAIL_HOST"] != "lerd-mailpit" {
		t.Errorf("bare hosts were rewritten: redis=%q mail=%q", updates["REDIS_HOST"], updates["MAIL_HOST"])
	}
}

// nginx serves the domain on 443, so the container port carried over from the
// old value would point at a port nothing listens on.
func TestApplyServiceDomainEnv_DropsThePortTheSchemeImplies(t *testing.T) {
	updates := map[string]string{"AWS_ENDPOINT": "http://lerd-rustfs:9000/"}
	applyServiceDomainEnv(updates,
		map[string]string{"rustfs": "rustfs.test"},
		map[string]int{"rustfs": 9000}, nil)
	if got := updates["AWS_ENDPOINT"]; got != "https://rustfs.test/" {
		t.Errorf("AWS_ENDPOINT = %q, want https://rustfs.test/", got)
	}
}

func TestApplyServiceDomainEnv_NoDomainsLeavesEverythingAlone(t *testing.T) {
	updates := map[string]string{"AWS_ENDPOINT": "http://lerd-rustfs:9000"}
	applyServiceDomainEnv(updates, nil, nil, nil)
	if updates["AWS_ENDPOINT"] != "http://lerd-rustfs:9000" {
		t.Errorf("value changed with no domains configured: %q", updates["AWS_ENDPOINT"])
	}
}

// A value naming a host that only shares the port must keep its own port.
func TestApplyServiceDomainEnv_LeavesOtherHostsPortsIntact(t *testing.T) {
	updates := map[string]string{"OTHER_URL": "http://example.test:9000/thing"}
	applyServiceDomainEnv(updates,
		map[string]string{"rustfs": "rustfs.test"},
		map[string]int{"rustfs": 9000}, nil)
	if got := updates["OTHER_URL"]; got != "http://example.test:9000/thing" {
		t.Errorf("OTHER_URL = %q, want it untouched", got)
	}
}

// A port clash moves the service, but the env_vars its preset hands sites still
// name the port it declares, which nothing listens on any more.
func TestFollowPublishedPort(t *testing.T) {
	ports := []string{"3002:3000"}
	moved := config.ServiceConfig{PublishedPort: 3004}
	cases := map[string]string{
		"http://localhost:3002/script.js": "http://localhost:3004/script.js",
		"http://127.0.0.1:3002":           "http://127.0.0.1:3004",
		"http://lerd-umami:3000":          "http://lerd-umami:3000",
		"http://localhost:9999/x":         "http://localhost:9999/x",
	}
	for in, want := range cases {
		if got := followPublishedPort(in, ports, moved); got != want {
			t.Errorf("followPublishedPort(%q) = %q, want %q", in, got, want)
		}
	}
	if got := followPublishedPort("http://localhost:3002", ports, config.ServiceConfig{}); got != "http://localhost:3002" {
		t.Errorf("an unmoved service changed: %q", got)
	}
}

// A domain serves one container port. A service with two, smocker's mock
// server beside its admin API, must keep the other one on the container name,
// or the app's outgoing calls land on the admin API.
func TestApplyServiceDomainEnv_OnlyThePortTheDomainServes(t *testing.T) {
	updates := map[string]string{
		"SMOCKER_URL":       "http://lerd-smocker:8080",
		"SMOCKER_ADMIN_URL": "http://lerd-smocker:8081",
	}
	applyServiceDomainEnv(updates,
		map[string]string{"smocker": "smocker.test"},
		nil,
		map[string]int{"smocker": 8081})
	if got := updates["SMOCKER_URL"]; got != "http://lerd-smocker:8080" {
		t.Errorf("SMOCKER_URL = %q, want the mock server left on the container", got)
	}
	if got := updates["SMOCKER_ADMIN_URL"]; got != "https://smocker.test" {
		t.Errorf("SMOCKER_ADMIN_URL = %q, want https://smocker.test", got)
	}
}

// A service's name can begin another's, so the container host has to end where
// the name does.
func TestApplyServiceDomainEnv_MatchesTheWholeContainerName(t *testing.T) {
	updates := map[string]string{
		"DB_URL":  "http://lerd-postgres-pgvector:5432/app",
		"APP_URL": "http://lerd-postgres/x",
	}
	applyServiceDomainEnv(updates, map[string]string{"postgres": "postgres.test"}, nil, nil)
	if got := updates["DB_URL"]; got != "http://lerd-postgres-pgvector:5432/app" {
		t.Errorf("DB_URL = %q, another service's container was rewritten", got)
	}
	if got := updates["APP_URL"]; got != "https://postgres.test/x" {
		t.Errorf("APP_URL = %q, want https://postgres.test/x", got)
	}
}
