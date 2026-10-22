package podman

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/geodro/lerd/internal/platform"
)

// bridgeRun is the podman call PinBridgeMAC makes, a seam for tests.
var bridgeRun = Run

// PinBridgeMAC fixes a network's bridge to the MAC it has now. Before netavark
// 1.9 a Linux bridge takes the lowest MAC among its ports, so a service
// container starting or stopping could move it, and containers that cached the
// old one lost the gateway, aardvark-dns with it, for up to 30s: nginx then
// failed to resolve the PHP pool and answered 502 (Ubuntu 24.04, WSL). An
// address set by hand stops the bridge following its ports. A no-op on newer
// netavark, which pins it itself, and when the bridge is not up yet, since it
// only exists while a container is on it.
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
		return nil
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
