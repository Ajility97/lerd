package ui

import (
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteBackupContent(t *testing.T) {
	serve := func(data []byte, err error) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		writeBackupContent(rec, httptest.NewRequest(http.MethodGet, "/x", nil), data, err)
		return rec
	}
	if rec := serve(nil, fs.ErrNotExist); rec.Code != http.StatusNotFound {
		t.Errorf("missing backup = %d, want 404", rec.Code)
	}
	if rec := serve(nil, errors.New("denied")); rec.Code != http.StatusInternalServerError {
		t.Errorf("unreadable backup = %d, want 500", rec.Code)
	}
	rec := serve([]byte("worker_processes 2;"), nil)
	if rec.Code != http.StatusOK || rec.Body.String() != "worker_processes 2;" || rec.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("backup = %d %q %q", rec.Code, rec.Body.String(), rec.Header().Get("Cache-Control"))
	}
}
