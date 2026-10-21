package tui

import (
	"sync"
	"time"

	"github.com/geodro/lerd/internal/lanip"
)

// The LAN IP lookup fires on every render when a site has LAN share on, so
// we cache it for a few seconds. The underlying net.Dial trick is cheap but
// re-running it every 100 ms of key input is wasteful when the routing
// table almost never changes between frames.
var lanIPCache struct {
	sync.Mutex
	value string
	at    time.Time
}

const lanIPTTL = 5 * time.Second

// primaryLANIP returns the host's primary outbound IPv4 address, or "" when
// detection fails. Mirrors cli.detectPrimaryLANIP so the TUI and the web UI
// agree on what URL the user should share — but kept local to avoid an
// import cycle between internal/tui and internal/cli.
func primaryLANIP() string {
	lanIPCache.Lock()
	defer lanIPCache.Unlock()
	if lanIPCache.value != "" && time.Since(lanIPCache.at) < lanIPTTL {
		return lanIPCache.value
	}

	ip := lanip.Primary()
	if ip != "" {
		lanIPCache.value = ip
		lanIPCache.at = time.Now()
	}
	return ip
}
