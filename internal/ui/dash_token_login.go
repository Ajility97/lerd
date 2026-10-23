package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/geodro/lerd/internal/config"
)

// dashTokenLoginFor is config.DashboardTokenLoginFor; tests swap it.
var dashTokenLoginFor = config.DashboardTokenLoginFor

var dashTokenClient = &http.Client{Timeout: 10 * time.Second}

// serveDashTokenLogin logs in through the dashboard's own API and redirects the
// frame to the page that takes the token. Any failure redirects to the plain
// dashboard, whose own login form still works.
func serveDashTokenLogin(w http.ResponseWriter, r *http.Request, svc *config.CustomService, login *config.DashboardTokenLogin, body map[string]string) {
	target, err := url.Parse(resolveDashboardURL(svc, loadServicesMap()))
	if err != nil || target.Host == "" || !isLoopbackTarget(target.Hostname()) {
		http.Error(w, "invalid dashboard URL for "+svc.Name, http.StatusBadGateway)
		return
	}
	if !serveWhileWaking(w, r, svc.Name, target) {
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	origin := target.Scheme + "://" + target.Host
	token, err := fetchDashToken(origin+login.Request, body, login.Token)
	if err != nil {
		http.Redirect(w, r, target.String(), http.StatusFound)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, strings.TrimSuffix(dashProxyPath(svc.Name), "/"))
	if path == "" {
		path = "/"
	}
	if r.URL.RawQuery != "" {
		path += "?" + r.URL.RawQuery
	}
	open := strings.NewReplacer("{token}", url.QueryEscape(token), "{path}", url.QueryEscape(path)).Replace(login.Open)
	http.Redirect(w, r, origin+open, http.StatusFound)
}

func fetchDashToken(endpoint string, body map[string]string, field string) (string, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	resp, err := dashTokenClient.Post(endpoint, "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login returned %s", resp.Status)
	}
	var out map[string]any
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&out); err != nil {
		return "", err
	}
	token, _ := out[field].(string)
	if token == "" {
		return "", fmt.Errorf("login response carries no %q", field)
	}
	return token, nil
}
