//go:build !darwin

package cli

import "fmt"

// MachineMemoryStatus is empty off macOS: there is no Podman Machine VM whose
// memory lerd sizes.
func MachineMemoryStatus() (MachineMemory, error) { return MachineMemory{}, nil }

func ResizeMachineMemory(int64, func(StartEvent), ...string) error {
	return fmt.Errorf("resizing the Podman Machine is only supported on macOS")
}
