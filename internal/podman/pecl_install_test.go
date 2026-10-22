package podman

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// peclInstallScript rebuilds lerd-pecl-install from the printf lines the
// Containerfile writes it with, so the test runs the script the image runs.
func peclInstallScript(t *testing.T) string {
	t.Helper()
	tmpl, err := GetQuadletTemplate("lerd-php-fpm.Containerfile")
	if err != nil {
		t.Fatal(err)
	}
	start := strings.Index(tmpl, "RUN printf '%s\\n' \\\n")
	end := strings.Index(tmpl, "> /usr/local/bin/lerd-pecl-install")
	if start < 0 || end < start {
		t.Fatal("lerd-pecl-install printf block not found in the Containerfile")
	}
	var lines []string
	for _, l := range strings.Split(tmpl[start:end], "\n")[1:] {
		l = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(l), "\\"))
		if l == "" {
			continue
		}
		lines = append(lines, strings.TrimSuffix(strings.TrimPrefix(l, "'"), "'"))
	}
	return strings.Join(lines, "\n") + "\n"
}

// runPeclInstall runs the wrapper against stub php, pecl, wget, sha256sum, mv
// and build tools, and returns which installers it called in order and how it exited.
func runPeclInstall(t *testing.T, spec string, env ...string) (string, error) {
	t.Helper()
	if _, err := os.Stat("/usr/local/bin/pie"); err == nil {
		t.Skip("a real /usr/local/bin/pie skips the download path under test")
	}
	dir := t.TempDir()
	stubs := map[string]string{
		"php":       `if [ "$1" = "-r" ]; then echo "$PHP_ID"; exit 0; fi; echo "pie $*" >> "$LOG"; exit "${PIE_RC:-0}"`,
		"pecl":      `echo "pecl $*" >> "$LOG"; exit "${PECL_RC:-0}"`,
		"wget":      `for url; do :; done; echo "wget $url" >> "$LOG"; [ -z "$TARBALL" ] || cat "$TARBALL"`,
		"phpize":    `echo phpize >> "$LOG"`,
		"make":      `echo "make $*" >> "$LOG"`,
		"nproc":     `echo 2`,
		"sha256sum": `cat >/dev/null; exit "${SHA_RC:-0}"`,
		"mv":        `exit 0`,
	}
	for name, body := range stubs {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	script := filepath.Join(dir, "lerd-pecl-install")
	if err := os.WriteFile(script, []byte(peclInstallScript(t)), 0o755); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(dir, "calls.log")
	cmd := exec.Command("sh", script, spec)
	cmd.Env = append([]string{"PATH=" + dir + ":/usr/bin:/bin", "LOG=" + log}, env...)
	err := cmd.Run()
	data, _ := os.ReadFile(log)
	return strings.ReplaceAll(strings.TrimSpace(string(data)), "\n", " | "), err
}

// installCalls runs the wrapper and fails the test unless it succeeded.
func installCalls(t *testing.T, spec string, env ...string) string {
	t.Helper()
	calls, err := runPeclInstall(t, spec, env...)
	if err != nil {
		t.Fatalf("lerd-pecl-install %s: %v (calls %q)", spec, err, calls)
	}
	return calls
}

// pieFetch is the stub log line for the wrapper downloading PIE itself.
const pieFetch = "wget https://github.com/php/pie/releases/download/1.5.2/pie.phar | "

func TestPeclInstallUsesPIEFirst(t *testing.T) {
	got := installCalls(t, "pcov", "PHP_ID=80400")
	if got != pieFetch+"pie /usr/local/bin/pie install --skip-enable-extension pecl/pcov" {
		t.Errorf("calls = %q, want PIE only", got)
	}
}

func TestPeclInstallFallsBackToPecl(t *testing.T) {
	cases := map[string]struct {
		spec string
		env  []string
		want string
	}{
		"PIE build fails":      {"xdebug", []string{"PHP_ID=80400", "PIE_RC=1"}, pieFetch + "pie /usr/local/bin/pie install --skip-enable-extension xdebug/xdebug | pecl install xdebug"},
		"PIE checksum fails":   {"redis", []string{"PHP_ID=80400", "SHA_RC=1"}, pieFetch + "pecl install redis"},
		"PHP below 8.1":        {"xdebug-3.3.2", []string{"PHP_ID=80030"}, "pecl install xdebug-3.3.2"},
		"no PIE package known": {"yaml", []string{"PHP_ID=80400"}, "pecl install yaml"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if got := installCalls(t, c.spec, c.env...); got != c.want {
				t.Errorf("calls = %q, want %q", got, c.want)
			}
		})
	}
}

func TestPeclInstallAsksPIEForTheIgbinaryRC(t *testing.T) {
	got := installCalls(t, "igbinary", "PHP_ID=80400")
	if got != pieFetch+"pie /usr/local/bin/pie install --skip-enable-extension igbinary/igbinary:^3.2@RC" {
		t.Errorf("calls = %q, want PIE with the RC allowed", got)
	}
}

// sourceTarball packs a release tarball whose configure only logs, laid out
// the way pecl.php.net ships one: a single versioned directory at the root.
func sourceTarball(t *testing.T) string {
	t.Helper()
	src := t.TempDir()
	pkg := filepath.Join(src, "igbinary-3.2.17RC1")
	if err := os.Mkdir(pkg, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkg, "configure"), []byte("#!/bin/sh\necho configure >> \"$LOG\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	tgz := filepath.Join(t.TempDir(), "igbinary.tgz")
	if out, err := exec.Command("tar", "czf", tgz, "-C", src, "igbinary-3.2.17RC1").CombinedOutput(); err != nil {
		t.Fatalf("tar: %v\n%s", err, out)
	}
	return tgz
}

func TestPeclInstallBuildsTheTarballWhenPeclFails(t *testing.T) {
	got := installCalls(t, "igbinary", "PHP_ID=70433", "PECL_RC=1", "TARBALL="+sourceTarball(t))
	if want := "pecl install igbinary | wget https://pecl.php.net/get/igbinary | phpize | configure | make -j2 | make install"; got != want {
		t.Errorf("calls = %q, want %q", got, want)
	}
}

func TestPeclInstallFailsWhenTheTarballDoesNot(t *testing.T) {
	got, err := runPeclInstall(t, "igbinary", "PHP_ID=70433", "PECL_RC=1")
	if err == nil {
		t.Errorf("an empty download must fail the install, calls %q", got)
	}
}
