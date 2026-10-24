package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/geodro/lerd/internal/plugins"
	"github.com/geodro/lerd/internal/push"
)

// pluginMessageRequestFor builds a loopback POST carrying the echo plugin's token.
func pluginMessageRequestFor(t *testing.T, body string) *http.Request {
	t.Helper()
	token, err := plugins.Token("echo")
	if err != nil {
		t.Fatal(err)
	}
	req := localRequest(http.MethodPost, "/api/notifications/plugin", body)
	req.Header.Set(plugins.TokenHeader, token)
	return req
}

func resetPluginMessageRate(t *testing.T) {
	t.Helper()
	pluginMessageRate.Lock()
	pluginMessageRate.sent = map[string][]time.Time{}
	pluginMessageRate.Unlock()
}

// The message is dispatched under the plugin the token belongs to, titled with
// its name so it never passes for lerd's own.
func TestPluginMessage_DispatchesUnderTokensPlugin(t *testing.T) {
	notifyGateConfigDir(t)
	installAPIPlugin(t)
	resetPluginMessageRate(t)
	ch := broker.add()
	defer broker.remove(ch)

	rec := httptest.NewRecorder()
	handlePluginMessage(rec, pluginMessageRequestFor(t, `{"title":"Hosts out of date","body":"Run sync.","url":"#sites/x","tag":"hosts"}`))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204 (%s)", rec.Code, rec.Body)
	}

	select {
	case msg := <-ch:
		var n push.Notification
		if err := json.Unmarshal(msg.Notification, &n); err != nil {
			t.Fatal(err)
		}
		if n.Kind != "plugin_message" || n.Title != "echo: Hosts out of date" || n.Body != "Run sync." ||
			n.URL != "#sites/x" || n.Tag != "lerd-plugin-message:echo:hosts" || n.Params["plugin"] != "echo" {
			t.Errorf("notification = %+v", n)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected the message to be broadcast")
	}
}

func TestPluginMessage_RefusesWithoutValidToken(t *testing.T) {
	notifyGateConfigDir(t)
	installAPIPlugin(t)
	for _, token := range []string{"", "bogus"} {
		req := localRequest(http.MethodPost, "/api/notifications/plugin", `{"title":"hi"}`)
		if token != "" {
			req.Header.Set(plugins.TokenHeader, token)
		}
		rec := httptest.NewRecorder()
		handlePluginMessage(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("token %q: status = %d, want 401", token, rec.Code)
		}
	}
}

func TestPluginMessage_RefusesOffLoopback(t *testing.T) {
	notifyGateConfigDir(t)
	installAPIPlugin(t)
	req := pluginMessageRequestFor(t, `{"title":"hi"}`)
	req.RemoteAddr = "192.168.1.42:54321"
	rec := httptest.NewRecorder()
	handlePluginMessage(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
}

// A field lerd does not know, such as a plugin name of its own choosing, is
// refused rather than ignored.
func TestPluginMessage_RefusesInvalidBodies(t *testing.T) {
	notifyGateConfigDir(t)
	installAPIPlugin(t)
	resetPluginMessageRate(t)
	for _, body := range []string{
		`{"title":"hi","plugin":"other"}`,
		`{"body":"no title"}`,
		`{"title":"   "}`,
		`{"title":"` + strings.Repeat("x", 101) + `"}`,
		`{"title":"hi","body":"` + strings.Repeat("x", 301) + `"}`,
		`{"title":"hi","url":"https://example.com"}`,
		`{"title":"hi","tag":"Not OK"}`,
		`not json`,
	} {
		rec := httptest.NewRecorder()
		handlePluginMessage(rec, pluginMessageRequestFor(t, body))
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", body, rec.Code)
		}
	}
}

// Limits count characters, not bytes, so a message in Danish or Chinese gets
// the same room as one in English.
func TestPluginMessage_LimitsCountCharacters(t *testing.T) {
	m := pluginMessageRequest{Title: strings.Repeat("æ", 100), Body: strings.Repeat("ø", 300)}
	if msg := m.validate(); msg != "" {
		t.Errorf("validate = %q, want valid", msg)
	}
}

func TestPluginMessage_RateLimitedPerPlugin(t *testing.T) {
	resetPluginMessageRate(t)
	now := time.Now()
	for i := 0; i < pluginMessageLimit; i++ {
		if !allowPluginMessage("echo", now) {
			t.Fatalf("message %d refused, want allowed", i+1)
		}
	}
	if allowPluginMessage("echo", now) {
		t.Error("message past the limit allowed")
	}
	if !allowPluginMessage("other", now) {
		t.Error("another plugin refused, want its own limit")
	}
	if !allowPluginMessage("echo", now.Add(pluginMessageWindow)) {
		t.Error("message after the window refused")
	}
}

// An empty tag is no way out: the service worker tags it lerd-<kind>, so every
// untagged message would replace the last in Web Push. Each gets one of its own.
func TestPluginMessage_UntaggedMessagesStayApart(t *testing.T) {
	a := pluginMessageNotification("echo", pluginMessageRequest{Title: "hi"}).Tag
	b := pluginMessageNotification("echo", pluginMessageRequest{Title: "hi"}).Tag
	if a == "" || b == "" || a == b {
		t.Errorf("untagged messages got %q and %q, want two different tags", a, b)
	}
	if tagged := pluginMessageNotification("echo", pluginMessageRequest{Title: "hi", Tag: "1"}).Tag; tagged == a || tagged == b {
		t.Errorf("an untagged message took the tag a plugin can ask for: %q", tagged)
	}
}

// The plugin and its tag are kept apart, so no two plugins share a tag.
func TestPluginMessage_TagsNeverCollideAcrossPlugins(t *testing.T) {
	a := pluginMessageNotification("foo", pluginMessageRequest{Title: "x", Tag: "bar-baz"}).Tag
	b := pluginMessageNotification("foo-bar", pluginMessageRequest{Title: "x", Tag: "baz"}).Tag
	if a == b {
		t.Errorf("both plugins got tag %q", a)
	}
}

// A message at both limits written as \u escapes, as Python's json.dumps sends
// emoji, is accepted.
func TestPluginMessage_AcceptsEscapedMessageAtLimits(t *testing.T) {
	notifyGateConfigDir(t)
	installAPIPlugin(t)
	resetPluginMessageRate(t)
	emoji := `\ud83d\ude00`
	body := `{"title":"` + strings.Repeat(emoji, 100) + `","body":"` + strings.Repeat(emoji, 300) + `"}`
	rec := httptest.NewRecorder()
	handlePluginMessage(rec, pluginMessageRequestFor(t, body))
	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204 (%s)", rec.Code, rec.Body)
	}
}
