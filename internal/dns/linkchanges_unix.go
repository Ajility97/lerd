//go:build darwin || linux

package dns

import (
	"errors"
	"fmt"

	"golang.org/x/sys/unix"
)

// readLinkEvents reads the open, non-blocking kernel socket fd until done
// closes, signalling out (without blocking) for each read that relevant
// accepts; a nil relevant accepts every read. label names the socket in errors.
func readLinkEvents(fd int, out chan<- struct{}, done <-chan struct{}, label string, relevant func([]byte) bool) error {
	// stopped lets the interrupt goroutine exit when LinkChanges returns on its
	// own (e.g. a read error the supervisor will restart from), so a restart
	// loop doesn't accumulate one parked goroutine per attempt.
	stopped := make(chan struct{})
	gone := make(chan struct{})
	// Join the interrupt goroutine (via gone) before the deferred unix.Close(fd)
	// runs, so its unix.Shutdown(fd) can't land on an fd the runtime has already
	// closed and reassigned to a freshly-opened socket on a fast restart.
	defer func() { close(stopped); <-gone }()
	go func() {
		defer close(gone)
		select {
		case <-done:
		case <-stopped:
		}
		_ = unix.Shutdown(fd, unix.SHUT_RDWR)
	}()

	buf := make([]byte, 4096)
	for {
		n, err := unix.Read(fd, buf)
		if err != nil {
			if errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EWOULDBLOCK) {
				if waitErr := waitReadable(fd, done); waitErr != nil {
					return errUnlessDone(done, fmt.Errorf("%s poll: %w", label, waitErr))
				}
				continue
			}
			return errUnlessDone(done, fmt.Errorf("%s read: %w", label, err))
		}
		if n <= 0 {
			return errUnlessDone(done, fmt.Errorf("%s closed", label))
		}
		if relevant != nil && !relevant(buf[:n]) {
			continue
		}
		select {
		case out <- struct{}{}:
		default:
		}
	}
}

// waitReadable blocks until fd has data or done closes. The socket is
// non-blocking so a clean shutdown can interrupt the read by closing
// the socket, but we still need to park between bursts without spinning.
func waitReadable(fd int, done <-chan struct{}) error {
	pfd := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
	for {
		_, err := unix.Poll(pfd, 1000)
		if err != nil {
			if errors.Is(err, unix.EINTR) {
				select {
				case <-done:
					return errors.New("done")
				default:
					continue
				}
			}
			return err
		}
		select {
		case <-done:
			return errors.New("done")
		default:
		}
		if pfd[0].Revents&(unix.POLLIN|unix.POLLERR|unix.POLLHUP) != 0 {
			return nil
		}
	}
}

// errUnlessDone returns nil when done is closed, since a requested shutdown
// makes the in-flight socket read fail and that failure is expected. Otherwise
// it returns err so the caller logs a one-shot warning and falls back to its
// time-based poll, rather than silently going deaf to host network changes.
func errUnlessDone(done <-chan struct{}, err error) error {
	select {
	case <-done:
		return nil
	default:
		return err
	}
}
