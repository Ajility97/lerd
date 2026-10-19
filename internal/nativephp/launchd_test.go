package nativephp

import (
	"strings"
	"testing"
	"time"
)

func TestPlistCarriesEnvAndArgs(t *testing.T) {
	got := buildFPMPlist("lerd-native-php84",
		[]string{"/bin/php-fpm", "-y", "/conf/php-fpm.conf"},
		map[string]string{"PHP_INI_SCAN_DIR": "/a:/b"},
		"/logs/native84.log")

	for _, want := range []string{
		"<string>lerd-native-php84</string>",
		"<string>/bin/php-fpm</string>",
		"<string>-y</string>",
		"<key>PHP_INI_SCAN_DIR</key>",
		"<string>/a:/b</string>",
		"<key>KeepAlive</key>",
		"<key>RunAtLoad</key>",
		"/logs/native84.log",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("plist missing %q:\n%s", want, got)
		}
	}
}

// A path with an ampersand must not produce invalid XML that launchd rejects.
func TestPlistEscapesXML(t *testing.T) {
	got := buildFPMPlist("l", []string{"/bin/a&b"}, map[string]string{"K": "<v>"}, "/log")
	if strings.Contains(got, "/bin/a&b") || strings.Contains(got, "<v>") {
		t.Errorf("plist must escape XML metacharacters:\n%s", got)
	}
	if !strings.Contains(got, "&amp;") || !strings.Contains(got, "&lt;v&gt;") {
		t.Errorf("expected escaped forms:\n%s", got)
	}
}

// Env order must be stable or every regeneration rewrites the plist and
// needlessly restarts FPM.
func TestPlistIsDeterministic(t *testing.T) {
	env := map[string]string{"B": "2", "A": "1", "C": "3"}
	a := buildFPMPlist("l", []string{"/bin/x"}, env, "/log")
	b := buildFPMPlist("l", []string{"/bin/x"}, env, "/log")
	if a != b {
		t.Error("plist generation must be deterministic")
	}
}

func TestUnitLabel(t *testing.T) {
	if got := UnitLabel("8.4"); got != "lerd-native-php84" {
		t.Errorf("UnitLabel(8.4) = %q, want lerd-native-php84", got)
	}
	if UnitLabel("8.4") == UnitLabel("8.5") {
		t.Error("each version needs its own launchd label")
	}
	// Must not collide with the containerised FPM unit names.
	if strings.Contains(UnitLabel("8.4"), "-fpm") {
		t.Errorf("label %q is too close to the container unit name", UnitLabel("8.4"))
	}
}

// Ensure is called on every start, not just on a switch, so it has to be a
// no-op when the listener is already up. Bootstrapping an already-loaded job
// fails with "Bootstrap failed: 5", which surfaced as a warning per version on
// a perfectly healthy install.
// Loaded is not running. A job launchd holds but has not started must still be
// kicked, or Ensure reports success while nothing serves and every request
// 502s. This is the same distinction the worker start path gets wrong.
func TestLoadedButIdleJobIsStarted(t *testing.T) {
	var kicked []string
	err := bootLaunchdUnitWith("lerd-native-php84", "/tmp/x.plist",
		func(string) bool { return true },  // loaded
		func(string) bool { return false }, // but not running
		func(label, path string) error { return nil },
		func(label string) error { kicked = append(kicked, label); return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(kicked) != 1 {
		t.Errorf("a loaded-but-idle job must be started, got %v", kicked)
	}
}

func TestBootstrapIsSkippedWhenAlreadyLoaded(t *testing.T) {
	var booted []string
	loaded := true
	err := bootLaunchdUnitWith("lerd-native-php84", "/tmp/x.plist",
		func(string) bool { return loaded },
		func(string) bool { return true }, // already running
		func(label, path string) error { booted = append(booted, label); return nil },
		func(string) error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(booted) != 0 {
		t.Errorf("an already-loaded job must not be bootstrapped again, got %v", booted)
	}

	loaded = false
	var kickedAfterBootstrap []string
	if err := bootLaunchdUnitWith("lerd-native-php84", "/tmp/x.plist",
		func(string) bool { return loaded },
		func(string) bool { return false }, // RunAtLoad did not start it
		func(label, path string) error { booted = append(booted, label); return nil },
		func(label string) error { kickedAfterBootstrap = append(kickedAfterBootstrap, label); return nil }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(kickedAfterBootstrap) != 1 {
		t.Errorf("a freshly bootstrapped job that did not start must be kicked, got %v", kickedAfterBootstrap)
	}
	if len(booted) != 1 {
		t.Errorf("an absent job must be bootstrapped, got %v", booted)
	}
}

// An ini change only reaches requests once FPM reloads, and Ensure is
// deliberately a no-op for an already-running listener, so the ini path needs
// a restart rather than an ensure.
func TestReloadRestartsARunningListener(t *testing.T) {
	var kicked []string
	err := reloadWith("lerd-native-php84",
		func(string) bool { return true },
		func(label string) error { kicked = append(kicked, label); return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(kicked) != 1 {
		t.Errorf("a running listener must be restarted, got %v", kicked)
	}
}

// bootout returns while launchd still holds the job, SIGTERMed and exiting.
// Booting it back in that window fails (bootstrap 5, kickstart 37) and left a
// reinstall with no FPM and every site on 502, so replacing waits it out.
func TestReplaceWaitsForTheOldJobToLeave(t *testing.T) {
	var bootedOut []string
	polls := 0
	replaceUnitWith("lerd-native-php84",
		func(label string) { bootedOut = append(bootedOut, label) },
		func(string) bool { polls++; return polls < 3 }, // gone on the third look
		func(time.Duration) {})
	if len(bootedOut) != 1 {
		t.Errorf("the old job must be booted out once, got %v", bootedOut)
	}
	if polls != 3 {
		t.Errorf("replace must wait until the job is gone, looked %d times", polls)
	}
}

func TestReplaceGivesUpOnAJobThatNeverLeaves(t *testing.T) {
	var waited time.Duration
	replaceUnitWith("lerd-native-php84", func(string) {},
		func(string) bool { return true },
		func(d time.Duration) { waited += d })
	if waited < unloadTimeout {
		t.Errorf("replace must wait the whole budget before going on, waited %v", waited)
	}
}
