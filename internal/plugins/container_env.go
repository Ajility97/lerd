package plugins

import "strings"

// ContainerAPIURL is lerd's API as a plugin script running inside a container
// reaches it; 127.0.0.1 there is the container itself.
const ContainerAPIURL = "http://host.containers.internal:7073"

// ContainerEnvArgs carries a plugin's context into a container exec, for the
// plugin scripts that run in one, such as PHP through lerd's php shim. It is
// nil outside a plugin. Values go by name, so podman reads them from lerd's
// own environment and the API token never appears in argv; the API URL is
// pointed at the host, and LERD_BIN is left out since no lerd runs in there.
func ContainerEnvArgs(environ []string) []string {
	inPlugin := false
	var names []string
	for _, kv := range environ {
		name, _, _ := strings.Cut(kv, "=")
		if !strings.HasPrefix(name, "LERD_") || name == "LERD_BIN" || name == "LERD_API_URL" {
			continue
		}
		inPlugin = inPlugin || name == "LERD_PLUGIN_NAME"
		names = append(names, name)
	}
	if !inPlugin {
		return nil
	}
	args := make([]string, 0, 2*len(names)+2)
	for _, n := range names {
		args = append(args, "--env", n)
	}
	return append(args, "--env", "LERD_API_URL="+ContainerAPIURL)
}
