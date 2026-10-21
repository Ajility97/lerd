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

// runPeclInstall runs the wrapper against stub php, pecl, wget, sha256sum and
// mv, and returns which installers it called in order.
func runPeclInstall(t *testing.T, spec string, env ...string) []string {
	t.Helper()
	if _, err := os.Stat("/usr/local/bin/pie"); err == nil {
		t.Skip("a real /usr/local/bin/pie skips the download path under test")
	}
	dir := t.TempDir()
	stubs := map[string]string{
		"php":       `if [ "$1" = "-r" ]; then echo "$PHP_ID"; exit 0; fi; echo "pie $*" >> "$LOG"; exit "${PIE_RC:-0}"`,
		"pecl":      `echo "pecl $*" >> "$LOG"`,
		"wget":      `exit 0`,
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
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("lerd-pecl-install %s: %v\n%s", spec, err, out)
	}
	data, _ := os.ReadFile(log)
	return strings.Fields(strings.ReplaceAll(strings.TrimSpace(string(data)), "\n", " | "))
}

func TestPeclInstallUsesPIEFirst(t *testing.T) {
	got := strings.Join(runPeclInstall(t, "pcov", "PHP_ID=80400"), " ")
	if got != "pie /usr/local/bin/pie install --skip-enable-extension pecl/pcov" {
		t.Errorf("calls = %q, want PIE only", got)
	}
}

func TestPeclInstallFallsBackToPecl(t *testing.T) {
	cases := map[string]struct {
		spec string
		env  []string
		want string
	}{
		"PIE build fails":      {"xdebug", []string{"PHP_ID=80400", "PIE_RC=1"}, "pie /usr/local/bin/pie install --skip-enable-extension xdebug/xdebug | pecl install xdebug"},
		"PIE checksum fails":   {"redis", []string{"PHP_ID=80400", "SHA_RC=1"}, "pecl install redis"},
		"PHP below 8.1":        {"xdebug-3.3.2", []string{"PHP_ID=80030"}, "pecl install xdebug-3.3.2"},
		"no PIE package known": {"yaml", []string{"PHP_ID=80400"}, "pecl install yaml"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if got := strings.Join(runPeclInstall(t, c.spec, c.env...), " "); got != c.want {
				t.Errorf("calls = %q, want %q", got, c.want)
			}
		})
	}
}
