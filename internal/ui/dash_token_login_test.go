package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/geodro/lerd/internal/config"
)

func tokenLoginUpstream(t *testing.T, token string) (*httptest.Server, *map[string]string) {
	got := map[string]string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/auth/login" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewDecoder(r.Body).Decode(&got)
		_ = json.NewEncoder(w).Encode(map[string]string{"token": token})
	}))
	t.Cleanup(srv.Close)
	return srv, &got
}

func serveTokenLogin(t *testing.T, dashboard, path string) *httptest.ResponseRecorder {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	svc := &config.CustomService{Name: "tokenapp", Dashboard: dashboard}
	login := &config.DashboardTokenLogin{Request: "/api/auth/login", Token: "token", Open: "/sso?token={token}&url={path}"}
	req := httptest.NewRequest("GET", "http://localhost:7073"+path, nil)
	rec := httptest.NewRecorder()
	serveDashTokenLogin(rec, req, svc, login, map[string]string{"username": "admin", "password": "pw"})
	return rec
}

func TestServeDashTokenLogin_RedirectsWithToken(t *testing.T) {
	srv, got := tokenLoginUpstream(t, "a+b/c")
	rec := serveTokenLogin(t, srv.URL, "/_svc/tokenapp/websites")
	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", rec.Code)
	}
	if (*got)["username"] != "admin" || (*got)["password"] != "pw" {
		t.Errorf("login body = %v", *got)
	}
	want := srv.URL + "/sso?token=" + url.QueryEscape("a+b/c") + "&url=" + url.QueryEscape("/websites")
	if loc := rec.Header().Get("Location"); loc != want {
		t.Errorf("Location = %q, want %q", loc, want)
	}
}

// A login that fails leaves the dashboard on its own form rather than an error.
func TestServeDashTokenLogin_FallsBackToDashboard(t *testing.T) {
	srv, _ := tokenLoginUpstream(t, "")
	rec := serveTokenLogin(t, srv.URL, "/_svc/tokenapp/")
	if loc := rec.Header().Get("Location"); rec.Code != http.StatusFound || loc != srv.URL {
		t.Errorf("got %d %q, want 302 to %s", rec.Code, loc, srv.URL)
	}
}

func TestServeDashTokenLogin_RefusesNonLoopback(t *testing.T) {
	rec := serveTokenLogin(t, "http://203.0.113.7:3000", "/_svc/tokenapp/")
	if rec.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want 502 for a non-loopback dashboard", rec.Code)
	}
}
