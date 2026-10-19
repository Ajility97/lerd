package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/geodro/lerd/internal/cli"
)

func withMachineMemorySeams(t *testing.T, st cli.MachineMemory, resize func(int64, func(cli.StartEvent), ...string) error) {
	t.Helper()
	oldStatus, oldResize := machineMemoryStatusFn, resizeMachineMemoryFn
	machineMemoryStatusFn = func() cli.MachineMemory { return st }
	resizeMachineMemoryFn = resize
	t.Cleanup(func() { machineMemoryStatusFn, resizeMachineMemoryFn = oldStatus, oldResize })
}

func TestMachineMemoryReportsKills(t *testing.T) {
	withMachineMemorySeams(t, cli.MachineMemory{Kills: []string{"mysqld"}, CurrentMiB: 4096, SuggestedMiB: 6144}, nil)

	rec := httptest.NewRecorder()
	handleMachineMemory(rec, httptest.NewRequest(http.MethodGet, "/api/machine/memory", nil))

	var got cli.MachineMemory
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Kills) != 1 || got.SuggestedMiB != 6144 {
		t.Errorf("got %+v, want the kills and the suggested size", got)
	}
}

// The resize restarts every unit, so lerd-ui must be left out of the start the
// same way the dashboard's start button leaves it out.
func TestMachineMemoryResizeSkipsTheUIUnit(t *testing.T) {
	var gotMiB int64
	var gotSkip []string
	withMachineMemorySeams(t, cli.MachineMemory{}, func(mib int64, emit func(cli.StartEvent), skip ...string) error {
		gotMiB, gotSkip = mib, skip
		emit(cli.StartEvent{Phase: "done"})
		return nil
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/machine/memory", strings.NewReader(`{"mib":6144}`))
	handleMachineMemory(rec, req)

	if gotMiB != 6144 || len(gotSkip) != 1 || gotSkip[0] != uiUnit {
		t.Errorf("resize got (%d, %v), want (6144, [%s])", gotMiB, gotSkip, uiUnit)
	}
	if !strings.Contains(rec.Body.String(), `"phase":"done"`) {
		t.Errorf("stream = %q, want the start events", rec.Body.String())
	}
}
