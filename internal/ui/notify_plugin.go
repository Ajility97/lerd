package ui

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/geodro/lerd/internal/plugins"
	"github.com/geodro/lerd/internal/push"
)

// pluginFailureNotification tells the user a plugin failed and takes them to
// the plugin manager, where its row offers to report it to the author.
func pluginFailureNotification(f plugins.Failure) push.Notification {
	first, _, _ := strings.Cut(f.Error, "\n")
	first = truncate(first, 200)
	return push.Notification{
		Kind:     "plugin_failed",
		TitleKey: "notify_plugin_failed_title",
		Title:    "Plugin " + f.Plugin + " failed",
		BodyKey:  "notify_plugin_failed_body",
		Body:     f.What + ": " + first,
		Params:   map[string]string{"plugin": f.Plugin, "what": f.What, "message": first},
		Tag:      "lerd-plugin-failed-" + f.Plugin,
		URL:      "#plugins",
		Data:     map[string]string{"plugin": f.Plugin},
		Urgency:  "normal",
		TTL:      3600,
	}
}

// notifyPluginFailures has failures in the dashboard's own process, from
// plugin endpoints and site commands, notify straight away.
func notifyPluginFailures() {
	plugins.SetFailureNotifier(func(f plugins.Failure) { dispatchNotification(pluginFailureNotification(f)) })
}

// handleInternalPluginFailed raises the notification for a failure another
// lerd process recorded, such as a hook the watcher ran.
func handleInternalPluginFailed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !isLoopbackRequest(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	var req struct {
		Plugin string `json:"plugin"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if f, ok := plugins.LastFailure(req.Plugin); ok {
		dispatchNotification(pluginFailureNotification(f))
	}
	w.WriteHeader(http.StatusNoContent)
}

// pluginMessageRequest is the body of POST /api/notifications/plugin.
type pluginMessageRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url"`
	Tag   string `json:"tag"`
}

const (
	pluginMessageTitleMax = 100
	pluginMessageBodyMax  = 300
	pluginMessageLimit    = 10
	pluginMessageWindow   = time.Minute
	// pluginMessageMaxBytes leaves room for a title and body at their limits
	// written entirely as \u escapes, the way Python's json.dumps sends emoji.
	pluginMessageMaxBytes = 64 << 10
)

var pluginMessageTag = regexp.MustCompile(`^[a-z0-9-]{1,64}$`)

// pluginMessageSeq numbers untagged messages, and the start time keeps the
// numbers from repeating a tag a browser still holds after lerd-ui restarts.
var (
	pluginMessageSeq   atomic.Uint64
	pluginMessageEpoch = strconv.FormatInt(time.Now().UnixNano(), 36)
)

// pluginMessageRate keeps when each plugin last sent a message, so a hook stuck
// in a loop cannot flood the user.
var pluginMessageRate = struct {
	sync.Mutex
	sent map[string][]time.Time
}{sent: map[string][]time.Time{}}

// allowPluginMessage reports whether plugin may send another message now, and
// counts it when it may.
func allowPluginMessage(plugin string, now time.Time) bool {
	pluginMessageRate.Lock()
	defer pluginMessageRate.Unlock()
	recent := pluginMessageRate.sent[plugin][:0]
	for _, t := range pluginMessageRate.sent[plugin] {
		if now.Sub(t) < pluginMessageWindow {
			recent = append(recent, t)
		}
	}
	if len(recent) >= pluginMessageLimit {
		pluginMessageRate.sent[plugin] = recent
		return false
	}
	pluginMessageRate.sent[plugin] = append(recent, now)
	return true
}

// validate reports what is wrong with the message, or "" when nothing is.
func (m pluginMessageRequest) validate() string {
	switch {
	case strings.TrimSpace(m.Title) == "":
		return "title is required"
	case utf8.RuneCountInString(m.Title) > pluginMessageTitleMax:
		return "title is longer than 100 characters"
	case utf8.RuneCountInString(m.Body) > pluginMessageBodyMax:
		return "body is longer than 300 characters"
	case m.URL != "" && !strings.HasPrefix(m.URL, "#"):
		return "url must be a dashboard link starting with #"
	case m.Tag != "" && !pluginMessageTag.MatchString(m.Tag):
		return "tag may only hold a-z, 0-9 and -, at most 64 characters"
	}
	return ""
}

// pluginMessageNotification builds the notification for a plugin's message.
// The title always names the plugin, so a plugin cannot pass for lerd itself.
func pluginMessageNotification(plugin string, m pluginMessageRequest) push.Notification {
	// The colon cannot appear in a plugin name or a tag, so two plugins never
	// share one. An untagged message gets a tag of its own, since an empty one
	// becomes lerd-<kind> in the service worker and would replace the last.
	tag := "lerd-plugin-message:" + plugin + ":" + m.Tag
	if m.Tag == "" {
		tag = "lerd-plugin-message-once:" + plugin + ":" + pluginMessageEpoch + "-" + strconv.FormatUint(pluginMessageSeq.Add(1), 36)
	}
	return push.Notification{
		Kind:     "plugin_message",
		TitleKey: "notify_plugin_message_title",
		Title:    plugin + ": " + m.Title,
		Body:     m.Body,
		Params:   map[string]string{"plugin": plugin, "title": m.Title},
		Tag:      tag,
		URL:      m.URL,
		Data:     map[string]string{"plugin": plugin},
		Urgency:  "normal",
		TTL:      3600,
	}
}

// handlePluginMessage lets a plugin tell the user something, with the token
// lerd hands its scripts. The plugin is the one the token belongs to, never a
// name from the body.
func handlePluginMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !isLoopbackRequest(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	plugin := plugins.PluginForToken(r.Header.Get(plugins.TokenHeader))
	if plugin == "" {
		http.Error(w, "a plugin token is required", http.StatusUnauthorized)
		return
	}
	p, err := plugins.Find(plugin)
	if err != nil {
		http.Error(w, "plugin "+plugin+" is not installed", http.StatusForbidden)
		return
	}
	switch plugins.NotificationState(p) {
	case plugins.NotificationsNone:
		http.Error(w, "plugin.yaml does not declare notifications: true", http.StatusForbidden)
		return
	case plugins.NotificationsPending, plugins.NotificationsDenied:
		http.Error(w, "the user has not allowed notifications from "+plugin, http.StatusForbidden)
		return
	}
	var req pluginMessageRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, pluginMessageMaxBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if msg := req.validate(); msg != "" {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	if !allowPluginMessage(plugin, time.Now()) {
		http.Error(w, "too many messages, at most 10 a minute", http.StatusTooManyRequests)
		return
	}
	dispatchNotification(pluginMessageNotification(plugin, req))
	w.WriteHeader(http.StatusNoContent)
}
