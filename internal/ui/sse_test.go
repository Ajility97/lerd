package ui

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// nonFlusher hides the recorder's Flush, the way a wrapped writer can.
type nonFlusher struct{ http.ResponseWriter }

func TestOpenEventStream(t *testing.T) {
	rec := httptest.NewRecorder()
	if _, ok := openEventStream(rec); !ok {
		t.Fatal("refused a flushing writer")
	}
	if got := rec.Header().Get("Content-Type"); got != "text/event-stream" {
		t.Errorf("Content-Type = %q", got)
	}
	if got := rec.Header().Get("X-Accel-Buffering"); got != "no" {
		t.Errorf("X-Accel-Buffering = %q, nginx would buffer the stream", got)
	}

	rec = httptest.NewRecorder()
	if _, ok := openEventStream(nonFlusher{rec}); ok || rec.Code != http.StatusInternalServerError {
		t.Errorf("non-flusher ok=%v code=%d, want refused with 500", ok, rec.Code)
	}
}
