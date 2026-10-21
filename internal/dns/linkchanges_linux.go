package dns

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// rtnetlink legacy multicast group masks. Using the pre-shifted RTMGRP_*
// constants (as opposed to position-based RTNLGRP_*) lets us OR them
// straight into SockaddrNetlink.Groups, which is the legacy 32-bit mask
// field, without per-group bit shifting.
const linkChangeGroups = unix.RTMGRP_LINK |
	unix.RTMGRP_IPV4_IFADDR |
	unix.RTMGRP_IPV6_IFADDR

// LinkChanges opens an rtnetlink multicast subscription and emits a
// struct{} on out every time the kernel reports a link or address state
// change. Message contents are intentionally discarded: callers
// re-fingerprint the host DNS environment on each settled event, so we
// only need the "something moved" signal. Returns nil after done closes,
// or a non-nil error if the netlink socket can't be opened or bound so
// the caller can log a one-shot warning before falling back to its
// time-based poll.
func LinkChanges(out chan<- struct{}, done <-chan struct{}) error {
	fd, err := unix.Socket(unix.AF_NETLINK,
		unix.SOCK_RAW|unix.SOCK_CLOEXEC|unix.SOCK_NONBLOCK,
		unix.NETLINK_ROUTE)
	if err != nil {
		return fmt.Errorf("rtnetlink socket: %w", err)
	}
	defer unix.Close(fd)

	if err := unix.Bind(fd, &unix.SockaddrNetlink{
		Family: unix.AF_NETLINK,
		Groups: linkChangeGroups,
	}); err != nil {
		return fmt.Errorf("rtnetlink bind: %w", err)
	}

	return readLinkEvents(fd, out, done, "rtnetlink", nil)
}
