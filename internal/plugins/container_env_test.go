package plugins

import (
	"strings"
	"testing"
)

func TestContainerEnvArgs(t *testing.T) {
	if got := ContainerEnvArgs([]string{"LERD_SITE=acme", "HOME=/h"}); got != nil {
		t.Errorf("outside a plugin nothing is forwarded, got %v", got)
	}
	got := strings.Join(ContainerEnvArgs([]string{
		"LERD_PLUGIN_NAME=hello", "LERD_SITE=acme", "LERD_API_TOKEN=secret",
		"LERD_API_URL=" + APIURL, "LERD_BIN=/home/u/.local/bin/lerd", "HOME=/h",
	}), " ")
	for _, want := range []string{"--env LERD_PLUGIN_NAME", "--env LERD_SITE", "--env LERD_API_TOKEN", "--env LERD_API_URL=" + ContainerAPIURL} {
		if !strings.Contains(got, want) {
			t.Errorf("args %q, want %q", got, want)
		}
	}
	if strings.Contains(got, "secret") {
		t.Error("values go by name only, so the token never lands in podman's argv")
	}
	if strings.Contains(got, "LERD_BIN") || strings.Contains(got, "HOME") {
		t.Errorf("the host's lerd binary means nothing inside the container, got %q", got)
	}
}
