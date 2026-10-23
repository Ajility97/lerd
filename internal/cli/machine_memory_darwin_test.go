//go:build darwin

package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestMachineListHelperProcess is the child the fake machineCmdContext spawns:
// it prints GO_MACHINE_LIST_OUT, or fails when that is unset.
func TestMachineListHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_MACHINE_LIST_HELPER") != "1" {
		return
	}
	out, ok := os.LookupEnv("GO_MACHINE_LIST_OUT")
	if !ok {
		os.Exit(1)
	}
	fmt.Print(out)
	os.Exit(0)
}

func fakeMachineList(t *testing.T, out *string) {
	t.Helper()
	prev := machineCmdContext
	t.Cleanup(func() { machineCmdContext = prev })
	machineCmdContext = func(ctx context.Context, args ...string) *exec.Cmd {
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestMachineListHelperProcess", "--")
		cmd.Env = append(os.Environ(), "GO_WANT_MACHINE_LIST_HELPER=1")
		if out != nil {
			cmd.Env = append(cmd.Env, "GO_MACHINE_LIST_OUT="+*out)
		}
		return cmd
	}
}

// A machine list that fails must not read as a VM with no kills.
func TestMachineMemoryStatusReportsAFailedMachineList(t *testing.T) {
	fakeMachineList(t, nil)

	if _, err := MachineMemoryStatus(); err == nil {
		t.Fatal("expected an error when the machine list fails")
	}
}

// An unreadable console log means the check could not run, not that it passed.
func TestMachineMemoryStatusReportsAMissingConsoleLog(t *testing.T) {
	out := "lerd*\n"
	fakeMachineList(t, &out)
	t.Setenv("TMPDIR", t.TempDir())

	if _, err := MachineMemoryStatus(); err == nil {
		t.Fatal("expected an error when the console log is missing")
	}
}

func TestMachineMemoryStatusReadsTheConsoleLog(t *testing.T) {
	out := "lerd*\n"
	fakeMachineList(t, &out)
	dir := t.TempDir()
	t.Setenv("TMPDIR", dir)
	if err := os.MkdirAll(filepath.Join(dir, "podman"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "podman", "lerd.log"), []byte("booted\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	st, err := MachineMemoryStatus()
	if err != nil || len(st.Kills) != 0 {
		t.Fatalf("got %+v, %v; want no kills and no error", st, err)
	}
}

// fakeMachineStates answers each inspect with the next state, holding the last.
func fakeMachineStates(t *testing.T, states ...string) {
	t.Helper()
	prevCmd, prevPoll := machineCmdContext, machineStatePoll
	t.Cleanup(func() { machineCmdContext, machineStatePoll = prevCmd, prevPoll })
	machineStatePoll = 10 * time.Millisecond
	i := 0
	machineCmdContext = func(ctx context.Context, args ...string) *exec.Cmd {
		out := ""
		if len(args) > 1 && args[1] == "inspect" {
			out = states[min(i, len(states)-1)]
			i++
		}
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestMachineListHelperProcess", "--")
		cmd.Env = append(os.Environ(), "GO_WANT_MACHINE_LIST_HELPER=1", "GO_MACHINE_LIST_OUT="+out)
		return cmd
	}
}

func TestStopMachineAndWaitReturnsOnceStopped(t *testing.T) {
	fakeMachineStates(t, "running\n", "stopping\n", "stopped\n")

	if err := stopMachineAndWait("lerd", 10*time.Second); err != nil {
		t.Fatalf("got %v, want nil", err)
	}
}

// A set on a VM that is still up fails with exit 125, so the wait must say so first.
func TestStopMachineAndWaitNamesTheStateItGaveUpOn(t *testing.T) {
	fakeMachineStates(t, "stopping\n")

	err := stopMachineAndWait("lerd", 300*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "stopping") {
		t.Fatalf("got %v, want an error naming the stopping state", err)
	}
}
