//go:build darwin

package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/geodro/lerd/internal/podman"
)

// MachineMemoryStatus reports the processes the VM's kernel killed for lack of
// memory since it booted, read from its serial console log on the host, so no
// call reaches the VM. The memory sizes are only looked up when there were kills.
func MachineMemoryStatus() MachineMemory {
	name := selectedMachineName()
	if name == "" {
		return MachineMemory{}
	}
	b, err := os.ReadFile(machineConsoleLog(name))
	if err != nil {
		return MachineMemory{}
	}
	st := MachineMemory{Kills: podman.ParseOOMKills(string(b))}
	if len(st.Kills) == 0 {
		return st
	}
	out, err := machineQuery("machine", "inspect", "--format", "{{.Resources.Memory}}", name)
	if err != nil {
		return st
	}
	st.CurrentMiB, _ = strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	st.SuggestedMiB = raisedVMMemoryMiB(st.CurrentMiB, hostMemoryGiB())
	return st
}

// machineConsoleLog is where podman keeps the VM's serial console, in the
// per-user temp dir. launchd jobs do not always carry TMPDIR, hence getconf.
func machineConsoleLog(name string) string {
	dir := os.Getenv("TMPDIR")
	if dir == "" {
		out, _ := exec.Command("getconf", "DARWIN_USER_TEMP_DIR").Output()
		dir = strings.TrimSpace(string(out))
	}
	return filepath.Join(dir, "podman", name+".log")
}

// ResizeMachineMemory restarts the VM with mib of memory, then runs the normal
// start so every container comes back with its port forwards.
func ResizeMachineMemory(mib int64, emit func(StartEvent), skip ...string) error {
	name := selectedMachineName()
	if name == "" {
		return fmt.Errorf("no Podman Machine found; run `lerd start` to create one")
	}
	if ceiling := int64(hostMemoryGiB()) * 1024 / 2; mib < 2048 || mib > ceiling {
		return fmt.Errorf("%d MB is outside 2048 to %d MB, half of this Mac's memory", mib, ceiling)
	}
	release, err := podman.LockLifecycle()
	if err != nil {
		return fmt.Errorf("lifecycle lock: %w", err)
	}
	if emit != nil {
		emit(StartEvent{Phase: "step", Step: "machine"})
	}
	_ = runMachineStreaming(machineStopTimeout, "machine", "stop", name)
	err = runMachineStreaming(machineQueryTimeout, "machine", "set", "--memory", strconv.FormatInt(mib, 10), name)
	release()
	if err != nil {
		return fmt.Errorf("podman machine set --memory: %w", err)
	}
	return RunStart(emit, skip...)
}
