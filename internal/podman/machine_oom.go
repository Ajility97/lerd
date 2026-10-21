package podman

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	kernelStamp = regexp.MustCompile(`\[\s*(\d+)\.\d+\]`)
	oomKill     = regexp.MustCompile(`Out of memory: Killed process \d+ \(([^)]+)\)`)
)

// ParseOOMKills returns the processes the VM's kernel killed for lack of memory
// since its last boot, read from the machine's serial console log. The log can
// span boots, so a kernel uptime stamp going backwards starts the count over.
func ParseOOMKills(log string) []string {
	var kills []string
	last := int64(-1)
	for _, line := range strings.Split(log, "\n") {
		m := kernelStamp.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		stamp, _ := strconv.ParseInt(m[1], 10, 64)
		if stamp < last {
			kills = nil
		}
		last = stamp
		if k := oomKill.FindStringSubmatch(line); k != nil {
			kills = append(kills, k[1])
		}
	}
	return kills
}
