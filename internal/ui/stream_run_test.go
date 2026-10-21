package ui

import (
	"context"
	"net/http/httptest"
	"os/exec"
	"testing"
	"time"
)

// A script's background child inherits the output pipes, so killing only the
// script leaves the readers, and the site's run lock, waiting on it forever.
func TestStreamRun_DeadlineStopsWhatTheScriptStarted(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	cmd := exec.Command("sh", "-c", "sleep 30 & wait")
	done := make(chan struct{})
	go func() {
		streamRun(httptest.NewRecorder(), ctx, cmd, false)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("streamRun must return once its deadline stops the script")
	}
}
