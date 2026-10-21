package cli

import (
	"fmt"
	"strings"
)

// recommendedVMMemoryMiB picks Podman Machine memory based on host RAM and
// the active worker runtime mode. 8 GB is the minimum host RAM tier; anything
// below (including detection failures) uses the 8 GB values.
//
// Container mode (workers each get a dedicated container): max 6144 MiB.
// Exec mode (workers share the FPM container): max 4096 MiB — a ⅔ ratio
// applied proportionally across all tiers.
//
//	Host RAM      Container   Exec
//	≤ 8 GB         3072 MiB  2048 MiB
//	9 – 31 GB      4096 MiB  3072 MiB
//	≥ 32 GB        6144 MiB  4096 MiB
//
// int64 matches strconv.ParseInt at the comparison site.
func recommendedVMMemoryMiB(hostMemoryGiB int, execMode bool) int64 {
	switch {
	case hostMemoryGiB >= 32:
		if execMode {
			return 4096
		}
		return 6144
	case hostMemoryGiB >= 9:
		if execMode {
			return 3072
		}
		return 4096
	default: // ≤ 8 GB or detection failed — 8 GB is the floor
		if execMode {
			return 2048
		}
		return 3072
	}
}

// raisedVMMemoryMiB is the size to offer a VM that ran out of memory: 2 GB
// more, but never past half the host's RAM so macOS keeps room of its own.
// 0 when there is nothing more to give or the host size is unknown.
func raisedVMMemoryMiB(currentMiB int64, hostMemoryGiB int) int64 {
	ceiling := int64(hostMemoryGiB) * 1024 / 2
	next := min(currentMiB+2048, ceiling)
	if next <= currentMiB {
		return 0
	}
	return next
}

// MachineMemory is the Podman Machine's out-of-memory record since it booted.
type MachineMemory struct {
	Kills        []string `json:"kills"`
	CurrentMiB   int64    `json:"current_mib,omitempty"`
	SuggestedMiB int64    `json:"suggested_mib,omitempty"` // 0: nothing more to give
}

// Summary says how often the VM ran out of memory and what the kernel killed.
func (st MachineMemory) Summary() string {
	return fmt.Sprintf("ran out of memory %d times since it booted (%s)", len(st.Kills), strings.Join(dedupeStrings(st.Kills), ", "))
}

// Hint is the command that gives the VM more memory, or what to do instead
// when the host has none left to give.
func (st MachineMemory) Hint() string {
	if st.CurrentMiB == 0 {
		return "could not read the VM's memory size; run `lerd machine memory` again"
	}
	if st.SuggestedMiB == 0 {
		return "this Mac has no more memory to give it; stop sites or workers you are not using"
	}
	return fmt.Sprintf("lerd machine memory %d", st.SuggestedMiB/1024)
}
