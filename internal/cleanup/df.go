package cleanup

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/geodro/lerd/internal/podman"
)

// `podman images` charges every image its whole chain, `podman system df`
// charges shared layers to nobody but counts build intermediates as sharers, so
// df is only the fallback for images creditFreedLayers cannot size.

// On macOS these describe the store inside the Podman Machine VM; getting the
// blocks back on the host is `lerd machine reclaim`, the VM disk only grows.

// readUniqueBytes and readStoreBytes are the seams tests override.
var (
	readUniqueBytes = podmanUniqueBytes
	readStoreBytes  = podmanStoreBytes
)

// podmanUniqueBytes maps short image ID to the bytes only that image holds.
// An empty map means podman could not answer, and every caller then falls back
// to the unshared sizes podman images reported.
func podmanUniqueBytes() map[string]int64 {
	out, err := podman.Run("system", "df", "-v")
	if err != nil {
		return nil
	}
	return parseUniqueBytes(out)
}

// podmanStoreBytes is the whole image store with layers counted once, or 0 when
// podman could not answer.
func podmanStoreBytes() int64 {
	out, err := podman.Run("system", "df", "--format", "json")
	if err != nil {
		return 0
	}
	return parseStoreBytes(out)
}

// parseUniqueBytes reads the image rows of `podman system df -v`. The columns
// around CREATED are fixed but CREATED itself is a variable-width phrase ("5
// days", "About a minute"), so a row is read from both ends: the id is the
// third field and the unique size the second from last.
func parseUniqueBytes(out string) map[string]int64 {
	unique := map[string]int64{}
	inImages := false
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "REPOSITORY") {
			inImages = true
			continue
		}
		if !inImages {
			continue
		}
		f := strings.Fields(line)
		if len(f) == 0 {
			break // the blank line after the last image row ends the section
		}
		if len(f) < 8 {
			continue
		}
		if b, ok := parseHumanSize(f[len(f)-2]); ok {
			unique[f[2]] = b
		}
	}
	return unique
}

// parseStoreBytes pulls the images row's RawSize out of `podman system df
// --format json`, which is the deduplicated total.
func parseStoreBytes(out string) int64 {
	var rows []struct {
		Type    string `json:"Type"`
		RawSize int64  `json:"RawSize"`
	}
	if err := json.Unmarshal([]byte(out), &rows); err != nil {
		return 0
	}
	for _, r := range rows {
		if r.Type == "Images" {
			return r.RawSize
		}
	}
	return 0
}

// sizeUnits are the decimal suffixes podman formats sizes with.
var sizeUnits = map[string]float64{"B": 1, "KB": 1e3, "MB": 1e6, "GB": 1e9, "TB": 1e12, "PB": 1e15}

// parseHumanSize reads one of podman's formatted sizes ("8.698MB") back into
// bytes. The rounding podman already applied is kept; every figure lerd shows
// is rounded again for display anyway.
func parseHumanSize(s string) (int64, bool) {
	s = strings.TrimSpace(s)
	i := strings.IndexFunc(s, func(r rune) bool { return r != '.' && (r < '0' || r > '9') })
	if i <= 0 {
		return 0, false
	}
	n, err := strconv.ParseFloat(s[:i], 64)
	if err != nil {
		return 0, false
	}
	mult, ok := sizeUnits[strings.ToUpper(s[i:])]
	if !ok {
		return 0, false
	}
	return int64(n * mult), true
}
