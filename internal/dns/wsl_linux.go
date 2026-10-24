package dns

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/geodro/lerd/internal/feedback"
	"github.com/geodro/lerd/internal/wsl"
)

// wslUpstreamDropin keeps WSL's DNS proxy as systemd-resolved's upstream. It is
// named for WSL rather than lerd on purpose: uninstall must leave it behind, or
// the distro is left with a resolver that has no upstream at all.
const wslUpstreamDropin = "/etc/systemd/resolved.conf.d/wsl-upstream.conf"

// Seams for tests.
var (
	isWSL             = wsl.IsWSL
	resolvConfPath    = "/etc/resolv.conf"
	resolvedIsRunning = func() bool {
		return exec.Command("systemctl", "is-active", "--quiet", "systemd-resolved").Run() == nil
	}
	// wslGeneratedResolvConf is WSL's own file, which /etc/resolv.conf links to
	// until something else takes it over.
	wslGeneratedResolvConf = "/mnt/wsl/resolv.conf"
	resolvedUnitPresent    = func() bool {
		return exec.Command("systemctl", "cat", "systemd-resolved").Run() == nil
	}
	upstreamDropinPresent = func() bool {
		_, err := os.Stat(wslUpstreamDropin)
		return err == nil
	}
	haveAptGet = func() bool {
		_, err := exec.LookPath("apt-get")
		return err == nil
	}
	runSudo = func(args ...string) error {
		if out, err := exec.Command("sudo", args...).CombinedOutput(); err != nil {
			return fmt.Errorf("%v: %w: %s", args, err, out)
		}
		return nil
	}
	writeRootFile = sudoWriteFile
)

// EnsureWSLResolver gives a WSL distro systemd-resolved with WSL's DNS proxy
// upstream, which .test inside the distro is routed through. Debian's WSL image
// ships without it, and Arch's leaves it off.
func EnsureWSLResolver() error {
	if !isWSL() {
		return nil
	}
	if !resolvedIsRunning() {
		if !resolvedUnitPresent() {
			if !haveAptGet() {
				return fmt.Errorf("systemd-resolved is not installed, and lerd routes .%s inside WSL through it: install systemd-resolved, then run `lerd install` again", ConfiguredTLD())
			}
			feedback.Sudo("Installing systemd-resolved so ." + ConfiguredTLD() + " can resolve inside WSL")
			// A fresh WSL image can carry no package lists at all.
			if err := runSudo("apt-get", "update", "-q"); err != nil {
				return err
			}
			if err := runSudo("apt-get", "install", "-y", "systemd-resolved"); err != nil {
				return err
			}
		}
		if !resolvedIsRunning() {
			feedback.Sudo("Starting systemd-resolved so ." + ConfiguredTLD() + " can resolve inside WSL")
			if err := runSudo("systemctl", "enable", "--now", "systemd-resolved"); err != nil {
				return err
			}
		}
	}
	return handResolvConfToResolved()
}

// WSLResolverMissing reports a WSL distro with no systemd-resolved running, where
// .test resolves on Windows but not inside the distro.
func WSLResolverMissing() bool { return isWSL() && !resolvedIsRunning() }

// wslHandoverNameservers returns WSL's DNS proxy addresses and search suffixes
// when resolv.conf still needs handing to systemd-resolved, nil when there is nothing to do.
func wslHandoverNameservers() (nameservers, search []string) {
	if !isWSL() || !resolvedIsRunning() {
		return nil, nil
	}
	cur, _ := os.ReadFile(resolvConfPath)
	if ns, s := wsl.GeneratedNameservers(string(cur)); ns != nil {
		return ns, s
	}
	if upstreamDropinPresent() {
		return nil, nil
	}
	// Installing resolved relinks resolv.conf to its stub and leaves it with no
	// upstream at all, but WSL still writes its own copy under /mnt/wsl.
	mnt, _ := os.ReadFile(wslGeneratedResolvConf)
	return wsl.GeneratedNameservers(string(mnt))
}

// handResolvConfToResolved points resolv.conf at systemd-resolved's stub: WSL's
// own file goes straight to its DNS proxy, past every .test route lerd sets.
// The proxy stays upstream, and generateResolvConf=false keeps WSL off the file.
func handResolvConfToResolved() error {
	ns, search := wslHandoverNameservers()
	if ns == nil {
		return nil
	}
	feedback.Sudo("Handing WSL's resolv.conf to systemd-resolved so ." + ConfiguredTLD() + " can resolve")
	if err := writeRootFile(wslUpstreamDropin, []byte(wsl.ResolvedUpstreamDropin(ns, search)), 0644); err != nil {
		return err
	}
	cur, _ := os.ReadFile("/etc/wsl.conf")
	if next, changed := wsl.EnsureSectionLine(string(cur), "network", "generateResolvConf", "generateResolvConf=false"); changed {
		if err := writeRootFile("/etc/wsl.conf", []byte(next), 0644); err != nil {
			return err
		}
	}
	for _, args := range [][]string{
		{"ln", "-sf", "/run/systemd/resolve/stub-resolv.conf", resolvConfPath},
		{"systemctl", "restart", "systemd-resolved"},
	} {
		if err := runSudo(args...); err != nil {
			return err
		}
	}
	return nil
}
