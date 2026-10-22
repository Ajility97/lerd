package podman

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/geodro/lerd/internal/platform"
)

// bridgeRun is the podman call PinBridgeMAC makes, a seam for tests.
var bridgeRun = Run

// PinBridgeMAC fixes the bridge MAC on netavark before 1.9, which lets it follow
// the lowest port MAC so a container start or stop cut DNS for up to 30s. Newer
// netavark, podman machine and a bridge not up yet are left alone.
func PinBridgeMAC(network string) error {
	if platform.Current.UsesMachineVM {
		return nil
	}
	backend, err := bridgeRun("info", "--format", "{{.Host.NetworkBackendInfo.Version}}")
	if err != nil {
		return err
	}
	moves, err := bridgeMACMoves(backend)
	if err != nil || !moves {
		return err
	}
	iface, err := bridgeRun("network", "inspect", network, "--format", "{{.NetworkInterface}}")
	if err != nil || iface == "" {
		return err
	}
	link, err := bridgeRun("unshare", "--rootless-netns", "ip", "-br", "link", "show", iface)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return nil
		}
		return err
	}
	f := strings.Fields(link)
	if len(f) < 3 || strings.Count(f[2], ":") != 5 {
		return fmt.Errorf("no MAC address in %q", link)
	}
	_, err = bridgeRun("unshare", "--rootless-netns", "ip", "link", "set", iface, "address", f[2])
	return err
}

// bridgeMACMoves reports whether the network backend lets a bridge's MAC follow
// its ports: netavark before 1.9. backend reads like "netavark 1.4.0"; another
// backend is not ours to touch.
func bridgeMACMoves(backend string) (bool, error) {
	v, ok := strings.CutPrefix(strings.TrimSpace(backend), "netavark ")
	if !ok {
		return false, nil
	}
	parts := strings.SplitN(v, ".", 3)
	if len(parts) < 2 {
		return false, fmt.Errorf("unreadable network backend version %q", backend)
	}
	major, errA := strconv.Atoi(parts[0])
	minor, errB := strconv.Atoi(parts[1])
	if errA != nil || errB != nil {
		return false, fmt.Errorf("unreadable network backend version %q", backend)
	}
	return major < 1 || major == 1 && minor < 9, nil
}
