package lanip

import (
	"net"
	"testing"
)

// Whatever the host's network, Primary answers a usable IPv4 or nothing:
// never loopback, never IPv6.
func TestPrimaryIsEmptyOrANonLoopbackIPv4(t *testing.T) {
	got := Primary()
	if got == "" {
		return
	}
	ip := net.ParseIP(got)
	if ip == nil || ip.To4() == nil || ip.IsLoopback() {
		t.Fatalf("Primary() = %q, want a non-loopback IPv4", got)
	}
}
