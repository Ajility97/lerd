package podman

import (
	"errors"
	"strings"
	"testing"

	"github.com/geodro/lerd/internal/platform"
)

// fakeBridgeRun answers the podman calls PinBridgeMAC makes and records
// every argv, so a test can see whether the address was set and to what.
func fakeBridgeRun(t *testing.T, iface, link string) *[]string {
	t.Helper()
	return fakeBridgeRunOn(t, "netavark 1.4.0", iface, link)
}

func fakeBridgeRunOn(t *testing.T, backend, iface, link string) *[]string {
	t.Helper()
	return fakeBridgeRunErr(t, backend, iface, link, errors.New(`Device "podman1" does not exist.`))
}

// fakeBridgeRunErr is fakeBridgeRunOn with the error the link lookup returns
// when link is empty.
func fakeBridgeRunErr(t *testing.T, backend, iface, link string, linkErr error) *[]string {
	t.Helper()
	var calls []string
	prev := bridgeRun
	bridgeRun = func(args ...string) (string, error) {
		joined := strings.Join(args, " ")
		calls = append(calls, joined)
		switch {
		case strings.HasPrefix(joined, "info"):
			return backend, nil
		case strings.HasPrefix(joined, "network inspect"):
			return iface, nil
		case strings.Contains(joined, "-br link show"):
			if link == "" {
				return "", linkErr
			}
			return link, nil
		}
		return "", nil
	}
	t.Cleanup(func() { bridgeRun = prev })
	return &calls
}

func linuxHost(t *testing.T) {
	t.Helper()
	prev := platform.Current
	platform.Current.UsesMachineVM = false
	t.Cleanup(func() { platform.Current = prev })
}

func TestPinBridgeMAC_SetsTheBridgeToItsCurrentAddress(t *testing.T) {
	linuxHost(t)
	calls := fakeBridgeRun(t, "podman1", "podman1          UP             12:5f:3d:3b:08:28 <BROADCAST,MULTICAST,UP,LOWER_UP>")

	if err := PinBridgeMAC("lerd"); err != nil {
		t.Fatalf("PinBridgeMAC: %v", err)
	}
	want := "unshare --rootless-netns ip link set podman1 address 12:5f:3d:3b:08:28"
	if last := (*calls)[len(*calls)-1]; last != want {
		t.Fatalf("last call = %q, want %q", last, want)
	}
}

// The bridge only exists while a container is on the network; with none there
// is nothing to pin and nothing to report.
func TestPinBridgeMAC_NoBridgeYetIsNotAnError(t *testing.T) {
	linuxHost(t)
	calls := fakeBridgeRun(t, "podman1", "")

	if err := PinBridgeMAC("lerd"); err != nil {
		t.Fatalf("PinBridgeMAC: %v", err)
	}
	for _, c := range *calls {
		if strings.Contains(c, "address") {
			t.Fatalf("set an address with no bridge up: %q", c)
		}
	}
}

func TestPinBridgeMAC_SkipsPodmanMachine(t *testing.T) {
	prev := platform.Current
	platform.Current.UsesMachineVM = true
	t.Cleanup(func() { platform.Current = prev })
	calls := fakeBridgeRun(t, "podman1", "podman1 UP 12:5f:3d:3b:08:28 <UP>")

	if err := PinBridgeMAC("lerd"); err != nil {
		t.Fatalf("PinBridgeMAC: %v", err)
	}
	if len(*calls) != 0 {
		t.Fatalf("ran podman inside a machine VM host: %v", *calls)
	}
}

func TestPinBridgeMAC_RefusesAnUnreadableLink(t *testing.T) {
	linuxHost(t)
	fakeBridgeRun(t, "podman1", "podman1 UP")

	if err := PinBridgeMAC("lerd"); err == nil {
		t.Fatal("want an error naming the unreadable link, got nil")
	}
}

// netavark 1.9 pins the bridge MAC itself, so newer hosts are left alone.
func TestPinBridgeMAC_LeavesNetavark19AndNewerAlone(t *testing.T) {
	linuxHost(t)
	for _, v := range []string{"netavark 1.9.0", "netavark 1.16.1", "netavark 2.1.0"} {
		calls := fakeBridgeRunOn(t, v, "podman1", "podman1 UP 12:5f:3d:3b:08:28 <UP>")
		if err := PinBridgeMAC("lerd"); err != nil {
			t.Fatalf("%s: %v", v, err)
		}
		if len(*calls) != 1 {
			t.Fatalf("%s: want only the version probe, got %v", v, *calls)
		}
	}
}

func TestPinBridgeMAC_PinsBelowNetavark19(t *testing.T) {
	linuxHost(t)
	for _, v := range []string{"netavark 1.4.0", "netavark 1.8.2"} {
		calls := fakeBridgeRunOn(t, v, "podman1", "podman1 UP 12:5f:3d:3b:08:28 <UP>")
		if err := PinBridgeMAC("lerd"); err != nil {
			t.Fatalf("%s: %v", v, err)
		}
		if last := (*calls)[len(*calls)-1]; !strings.HasSuffix(last, "address 12:5f:3d:3b:08:28") {
			t.Fatalf("%s: not pinned, last call %q", v, last)
		}
	}
}

func TestPinBridgeMAC_SkipsOtherBackends(t *testing.T) {
	linuxHost(t)
	calls := fakeBridgeRunOn(t, "cni 1.1.1", "podman1", "podman1 UP 12:5f:3d:3b:08:28 <UP>")
	if err := PinBridgeMAC("lerd"); err != nil {
		t.Fatalf("PinBridgeMAC: %v", err)
	}
	if len(*calls) != 1 {
		t.Fatalf("want only the version probe on CNI, got %v", *calls)
	}
}

func TestPinBridgeMAC_RefusesAnUnreadableNetavarkVersion(t *testing.T) {
	linuxHost(t)
	fakeBridgeRunOn(t, "netavark dev", "podman1", "podman1 UP 12:5f:3d:3b:08:28 <UP>")
	err := PinBridgeMAC("lerd")
	if err == nil || !strings.Contains(err.Error(), "netavark dev") {
		t.Fatalf("want an error naming the version, got %v", err)
	}
}

// Only a bridge that is not up yet is quiet; a lookup that could not run is
// reported, so lerd start and install can print why nothing was pinned.
func TestPinBridgeMAC_ReportsALookupThatCouldNotRun(t *testing.T) {
	linuxHost(t)
	fakeBridgeRunErr(t, "netavark 1.4.0", "podman1", "", errors.New(`exec: "ip": executable file not found in $PATH`))

	if err := PinBridgeMAC("lerd"); err == nil {
		t.Fatal("want the lookup error, got nil")
	}
}
