package plugins

import (
	"regexp"

	"github.com/geodro/lerd/internal/php"
	"github.com/geodro/lerd/internal/version"
)

// runningVersion is the lerd a plugin is judged against. A var so tests can
// pretend to be another release.
var runningVersion = version.Version

var releaseRe = regexp.MustCompile(`^v?\d+\.\d+`)

// Compatible reports whether this lerd meets a plugin's `lerd` constraint.
// A plugin states the oldest lerd it needs and no newest, so a new release
// never strands it; breaking the contract is what api_version is for. A build
// from an untagged commit has no version to judge and loads everything.
func Compatible(constraint string) bool {
	if constraint == "" || !releaseRe.MatchString(runningVersion) {
		return true
	}
	return php.Satisfies(runningVersion, constraint)
}
