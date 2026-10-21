package plugins

import (
	"strings"
	"testing"
)

func runningOn(t *testing.T, platform string) {
	t.Helper()
	orig := runningPlatform
	runningPlatform = platform
	t.Cleanup(func() { runningPlatform = orig })
}

func TestRunsHere(t *testing.T) {
	runningOn(t, "macos")
	for platforms, want := range map[string]bool{"": true, "macos": true, "linux,macos": true, "linux": false} {
		var list []string
		if platforms != "" {
			list = strings.Split(platforms, ",")
		}
		if got := RunsHere(list); got != want {
			t.Errorf("RunsHere(%v) on macos = %v, want %v", list, got, want)
		}
	}
}

func TestReadPlugin_RefusesAnotherPlatformByName(t *testing.T) {
	runningOn(t, "macos")
	dir := writePlugin(t, t.TempDir(), "p", "api_version: 1\nname: p\nplatforms: [linux]\n")
	_, err := ReadPlugin(dir)
	if err == nil || !strings.Contains(err.Error(), "linux") || !strings.Contains(err.Error(), "macos") {
		t.Fatalf("err = %v, want it to name the platforms it runs on and this one", err)
	}
}

// Windows runs lerd inside WSL, so a plugin for Windows is a plugin for linux;
// a value lerd does not know is refused rather than guessed at.
func TestReadPlugin_RefusesAnUnknownPlatform(t *testing.T) {
	dir := writePlugin(t, t.TempDir(), "p", "api_version: 1\nname: p\nplatforms: [linux, windows]\n")
	if _, err := ReadPlugin(dir); err == nil || !strings.Contains(err.Error(), `"windows"`) {
		t.Fatalf("err = %v, want the unknown platform named", err)
	}
}

func TestPendingUpdate_SkipsACommitForAnotherPlatform(t *testing.T) {
	isolate(t)
	repo, approved := gitRepo(t, "hello")
	serveIndex(t, Listing{Name: "hello", Repo: "file://" + repo, Commit: approved})
	if _, err := InstallListed("hello"); err != nil {
		t.Fatal(err)
	}
	runningOn(t, "macos")
	serveIndex(t, Listing{Name: "hello", Repo: "file://" + repo, Commit: strings.Repeat("b", 40), Platforms: []string{"linux"}})
	if err := refreshCatalog(); err != nil {
		t.Fatal(err)
	}
	if l, ok := PendingUpdate("hello"); ok {
		t.Errorf("an update that cannot run here must not be offered, got %+v", l)
	}
}
