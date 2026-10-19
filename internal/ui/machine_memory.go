package ui

import (
	"encoding/json"
	"net/http"

	"github.com/geodro/lerd/internal/cli"
)

// Seams so the handler can be exercised without a Podman Machine.
var (
	machineMemoryStatusFn = cli.MachineMemoryStatus
	resizeMachineMemoryFn = cli.ResizeMachineMemory
)

// handleMachineMemory reports the VM's out-of-memory kills since boot (GET), or
// restarts it with more memory and streams the start that follows (POST).
func handleMachineMemory(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, machineMemoryStatusFn())
	case http.MethodPost:
		var body struct {
			MiB int64 `json:"mib"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		writeLine, _ := startNDJSONStream(w, r)
		if err := resizeMachineMemoryFn(body.MiB, func(evt cli.StartEvent) { writeLine(evt) }, uiUnit); err != nil {
			writeLine(cli.StartEvent{Phase: "failed", Error: err.Error()})
		}
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
