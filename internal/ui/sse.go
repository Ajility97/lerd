package ui

import "net/http"

// openEventStream starts a server-sent event response, with nginx's proxy
// buffering off so each frame reaches the browser as it is written. It answers
// 500 itself when the writer cannot flush.
func openEventStream(w http.ResponseWriter) (http.Flusher, bool) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return nil, false
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	return flusher, true
}
