package plugins

import (
	"regexp"
	"strconv"

	"github.com/geodro/lerd/internal/version"
)

// runningVersion is the lerd a plugin is judged against. A var so tests can
// pretend to be another release.
var runningVersion = version.Version

var (
	releaseRe    = regexp.MustCompile(`^v?(\d+)\.(\d+)`)
	lowerBoundRe = regexp.MustCompile(`^>=\s*(\d+)\.(\d+)(\.\d+)?$`)
	// git describe past a tag, or with local changes: a dev build.
	devBuildRe = regexp.MustCompile(`-\d+-g[0-9a-f]+(-dirty)?$|-dirty$`)
)

// Compatible reports whether this lerd meets a plugin's `>=X.Y` lower bound, on
// major and minor; other forms are refused. An untagged or dev build loads all,
// since it already holds the next release. Parsed here to link no service code.
func Compatible(constraint string) bool {
	if constraint == "" {
		return true
	}
	need := lowerBoundRe.FindStringSubmatch(constraint)
	if need == nil {
		return false
	}
	have := releaseRe.FindStringSubmatch(runningVersion)
	if have == nil || devBuildRe.MatchString(runningVersion) {
		return true
	}
	if a, b := atoi(have[1]), atoi(need[1]); a != b {
		return a > b
	}
	return atoi(have[2]) >= atoi(need[2])
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
