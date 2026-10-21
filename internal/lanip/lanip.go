// Package lanip finds the host's LAN IPv4 address. It is a leaf package
// (stdlib only) so dns, podman, tui and ui can share it without a cycle.
package lanip

import "net"

// Primary returns the local IPv4 address the kernel would use to reach a
// public destination, falling back to the first up, non-loopback interface
// address. Empty when neither yields one. The dial sends no packets.
func Primary() string {
	if ip := dialProbe(); ip != "" {
		return ip
	}
	return interfaceProbe()
}

func dialProbe() string {
	conn, err := net.Dial("udp4", "1.1.1.1:80")
	if err != nil {
		return ""
	}
	defer conn.Close()
	addr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || addr == nil {
		return ""
	}
	return addr.IP.String()
}

func interfaceProbe() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if v4 := ipnet.IP.To4(); v4 != nil && !v4.IsLoopback() {
					return v4.String()
				}
			}
		}
	}
	return ""
}
