//go:build darwin

package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
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
