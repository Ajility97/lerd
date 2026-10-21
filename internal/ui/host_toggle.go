package ui

import (
	"encoding/json"
	"net/http"
)

// readHostToggle accepts a POST {"enable": bool} from a caller with
// host-action authority, writing the refusal itself when the request is not one.
func readHostToggle(w http.ResponseWriter, r *http.Request) (enable, ok bool) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return false, false
	}
	if !hasHostActionAuthority(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return false, false
	}
	var req struct {
		Enable bool `json:"enable"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return false, false
	}
	return req.Enable, true
}

// serveHostToggle serves a toggle whose whole job is set(enable), answering
// with the state set returns.
func serveHostToggle[T any](w http.ResponseWriter, r *http.Request, set func(bool) (T, error)) {
	enable, ok := readHostToggle(w, r)
	if !ok {
		return
	}
	res, err := set(enable)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, res)
}

// readEnabledBody accepts a POST {"enabled": bool} for a settings switch,
// writing the refusal itself when the request is not one.
func readEnabledBody(w http.ResponseWriter, r *http.Request) (enabled, ok bool) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return false, false
	}
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return false, false
	}
	return body.Enabled, true
}
