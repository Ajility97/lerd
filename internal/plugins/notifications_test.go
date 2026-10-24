package plugins

import (
	"testing"
)

// The answer is derived from the manifest and lerd's config, so an update that
// starts asking is pending on its own, an answer holds across updates, and a
// removal forgets it so a reinstall asks again.
func TestNotificationState_FollowsTheUsersAnswer(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	state := func() string {
		p, err := Find("hello")
		if err != nil {
			t.Fatal(err)
		}
		return NotificationState(p)
	}

	writePlugin(t, Dir(), "hello", "api_version: 1\nname: hello\n")
	if got := state(); got != NotificationsNone {
		t.Fatalf("without the key: %q, want none", got)
	}
	if err := SetNotifications("hello", true); err == nil {
		t.Fatal("a plugin that does not send notifications cannot be allowed to")
	}

	writePlugin(t, Dir(), "hello", "api_version: 1\nname: hello\nnotifications: true\n")
	if got := state(); got != NotificationsPending {
		t.Fatalf("after an update adds the key: %q, want pending", got)
	}
	if err := SetNotifications("hello", false); err != nil {
		t.Fatal(err)
	}
	writePlugin(t, Dir(), "hello", "api_version: 1\nname: hello\nversion: 0.2.0\nnotifications: true\n")
	if got := state(); got != NotificationsDenied {
		t.Fatalf("after a further update: %q, want the answer kept", got)
	}

	if err := Remove("hello"); err != nil {
		t.Fatal(err)
	}
	writePlugin(t, Dir(), "hello", "api_version: 1\nname: hello\nnotifications: true\n")
	if got := state(); got != NotificationsPending {
		t.Fatalf("after a reinstall: %q, want pending again", got)
	}
}
