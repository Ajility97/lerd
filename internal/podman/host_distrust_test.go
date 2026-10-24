package podman

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func withHostCAConf(t *testing.T, content string) {
	t.Helper()
	old := hostCAConfPath
	t.Cleanup(func() { hostCAConfPath = old })
	hostCAConfPath = filepath.Join(t.TempDir(), "ca-certificates.conf")
	if content != "" {
		if err := os.WriteFile(hostCAConfPath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// A root the host distrusts is distrusted in the image too, and the trust
// store is rebuilt after.
func TestHostDistrustBlock_CarriesHostDistrust(t *testing.T) {
	withHostCAConf(t, "# header\nmozilla/ACCVRAIZ1.crt\n!mozilla/emSign_Root_CA_-_G1.crt\n!mozilla/emSign_ECC_Root_CA_-_C3.crt\n")
	got := hostDistrustBlock()
	for _, want := range []string{
		`-e 's|^mozilla/emSign_Root_CA_-_G1\.crt$|!mozilla/emSign_Root_CA_-_G1.crt|'`,
		`-e 's|^mozilla/emSign_ECC_Root_CA_-_C3\.crt$|!mozilla/emSign_ECC_Root_CA_-_C3.crt|'`,
		"/etc/ca-certificates.conf && update-ca-certificates\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("block lacks %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "ACCVRAIZ1") {
		t.Errorf("a trusted root was carried over:\n%s", got)
	}
}

// No file, as on macOS, or nothing distrusted adds nothing to the build.
func TestHostDistrustBlock_EmptyWithoutDistrust(t *testing.T) {
	for _, content := range []string{"", "mozilla/ACCVRAIZ1.crt\n"} {
		withHostCAConf(t, content)
		if got := hostDistrustBlock(); got != "" {
			t.Errorf("content %q: block = %q, want none", content, got)
		}
	}
}

// Only entries shaped like a certificate path are copied, so the host file
// cannot inject anything into the RUN line.
func TestHostDistrustBlock_IgnoresUnsafeEntries(t *testing.T) {
	withHostCAConf(t, "!mozilla/x.crt'; rm -rf / #.crt\n!mozilla/$(id).crt\n!mozilla/ok.crt\n")
	got := hostDistrustBlock()
	if strings.Contains(got, "rm -rf") || strings.Contains(got, "$(") {
		t.Errorf("unsafe entry carried over:\n%s", got)
	}
	if !strings.Contains(got, "!mozilla/ok.crt") {
		t.Errorf("safe entry missing:\n%s", got)
	}
}
