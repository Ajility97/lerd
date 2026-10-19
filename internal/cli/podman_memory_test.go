package cli

import (
	"strings"
	"testing"
)

func TestRecommendedVMMemoryMiB(t *testing.T) {
	cases := []struct {
		name     string
		host     int
		execMode bool
		want     int64
	}{
		// container mode
		// container mode
		{"detection failed", 0, false, 3072},
		{"negative invalid", -1, false, 3072},
		{"low-end 4GB", 4, false, 3072},
		{"8GB MacBook", 8, false, 3072},
		{"9GB edge", 9, false, 4096},
		{"16GB laptop", 16, false, 4096},
		{"31GB upper mid", 31, false, 4096},
		{"32GB workstation", 32, false, 6144},
		{"64GB power user", 64, false, 6144},
		// exec mode: ⅔ of container tier (workers share FPM, no per-worker containers)
		{"detection failed exec", 0, true, 2048},
		{"negative invalid exec", -1, true, 2048},
		{"low-end 4GB exec", 4, true, 2048},
		{"8GB MacBook exec", 8, true, 2048},
		{"9GB edge exec", 9, true, 3072},
		{"16GB laptop exec", 16, true, 3072},
		{"31GB upper mid exec", 31, true, 3072},
		{"32GB workstation exec", 32, true, 4096},
		{"64GB power user exec", 64, true, 4096},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := recommendedVMMemoryMiB(c.host, c.execMode); got != c.want {
				t.Errorf("recommendedVMMemoryMiB(%d, execMode=%v) = %d, want %d", c.host, c.execMode, got, c.want)
			}
		})
	}
}

func TestRaisedVMMemoryMiB(t *testing.T) {
	cases := []struct {
		current int64
		hostGiB int
		want    int64
	}{
		{4096, 48, 6144}, // room to spare: 2 GB more
		{4096, 10, 5120}, // capped at half the host
		{8192, 16, 0},    // already at half the host
		{3072, 0, 0},     // host size unknown: offer nothing
	}
	for _, c := range cases {
		if got := raisedVMMemoryMiB(c.current, c.hostGiB); got != c.want {
			t.Errorf("raisedVMMemoryMiB(%d, %d) = %d, want %d", c.current, c.hostGiB, got, c.want)
		}
	}
}

// A failed size lookup must not read as "this Mac has no memory to give".
func TestMachineMemoryHintWhenTheSizeIsUnknown(t *testing.T) {
	st := MachineMemory{Kills: []string{"mysqld"}}
	if got := st.Hint(); !strings.Contains(got, "could not read") {
		t.Errorf("Hint() = %q, want the unknown-size advice", got)
	}
}
