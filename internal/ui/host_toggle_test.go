package ui

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func serveToggle(t *testing.T, method, remote, body string, set func(bool) (map[string]bool, error)) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, "/api/toggle", strings.NewReader(body))
	req.RemoteAddr = remote
	rec := httptest.NewRecorder()
	serveHostToggle(rec, req, set)
	return rec
}

func TestServeHostToggle(t *testing.T) {
	var got []bool
	set := func(on bool) (map[string]bool, error) {
		got = append(got, on)
		return map[string]bool{"on": on}, nil
	}
	local := "127.0.0.1:5555"

	if rec := serveToggle(t, http.MethodGet, local, `{"enable":true}`, set); rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET = %d, want 405", rec.Code)
	}
	if rec := serveToggle(t, http.MethodPost, "192.0.2.7:5555", `{"enable":true}`, set); rec.Code != http.StatusForbidden {
		t.Errorf("remote = %d, want 403", rec.Code)
	}
	if rec := serveToggle(t, http.MethodPost, local, `not json`, set); rec.Code != http.StatusBadRequest {
		t.Errorf("bad body = %d, want 400", rec.Code)
	}
	if len(got) != 0 {
		t.Fatalf("set ran on a refused request: %v", got)
	}

	rec := serveToggle(t, http.MethodPost, local, `{"enable":true}`, set)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"on":true`) {
		t.Errorf("enable = %d %s", rec.Code, rec.Body.String())
	}

	failing := func(bool) (map[string]bool, error) { return nil, errors.New("restart failed") }
	if rec := serveToggle(t, http.MethodPost, local, `{"enable":false}`, failing); rec.Code != http.StatusInternalServerError {
		t.Errorf("failing set = %d, want 500", rec.Code)
	}
}

func TestReadEnabledBody(t *testing.T) {
	read := func(method, body string) (bool, bool, int) {
		rec := httptest.NewRecorder()
		on, ok := readEnabledBody(rec, httptest.NewRequest(method, "/api/settings/x", strings.NewReader(body)))
		return on, ok, rec.Code
	}
	if _, ok, code := read(http.MethodGet, `{"enabled":true}`); ok || code != http.StatusMethodNotAllowed {
		t.Errorf("GET ok=%v code=%d", ok, code)
	}
	if _, ok, code := read(http.MethodPost, `{`); ok || code != http.StatusBadRequest {
		t.Errorf("bad body ok=%v code=%d", ok, code)
	}
	if on, ok, _ := read(http.MethodPost, `{"enabled":true}`); !ok || !on {
		t.Errorf("enable on=%v ok=%v", on, ok)
	}
}
