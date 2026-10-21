package ui

import (
	"net/http"
	"os"
)

// writeBackupContent answers a config editor's request for one backup's bytes:
// 404 when the backup is gone, 500 when it could not be read, the raw text
// otherwise.
func writeBackupContent(w http.ResponseWriter, r *http.Request, data []byte, err error) {
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "reading backup: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(data)
}
