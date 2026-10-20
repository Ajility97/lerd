package winshim

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

// ShimNames lists the shims in dir that get a .exe twin on Windows: the shell
// scripts lerd writes, not the binaries and version stamps beside them.
func ShimNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || strings.Contains(e.Name(), ".") {
			continue
		}
		f, err := os.Open(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		head := make([]byte, 2)
		n, _ := f.Read(head)
		f.Close()
		if string(head[:n]) == "#!" {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// SyncTwins makes php.exe and its siblings copies or links of lerd.exe (a .cmd
// fails, cmd.exe refuses a \\wsl.localhost folder). Stale twins go; a running
// one is renamed away.
func SyncTwins(dir string, names []string) error {
	lerdExe := filepath.Join(dir, "lerd.exe")
	want := map[string]bool{}
	for _, n := range names {
		want[n+".exe"] = true
	}
	stale, _ := filepath.Glob(filepath.Join(dir, "*.cmd"))
	exes, _ := filepath.Glob(filepath.Join(dir, "*.exe"))
	for _, f := range exes {
		if f != lerdExe && (!want[filepath.Base(f)] || !sameProgram(f, lerdExe)) {
			stale = append(stale, f)
		}
	}
	for _, f := range stale {
		if os.Remove(f) != nil {
			os.Remove(f + ".old")  //nolint:errcheck
			os.Rename(f, f+".old") //nolint:errcheck // still running; the next sync clears it
		}
	}
	for _, name := range names {
		dst := filepath.Join(dir, name+".exe")
		if _, err := os.Stat(dst); err == nil {
			continue
		}
		if os.Link(lerdExe, dst) == nil {
			continue
		}
		b, err := os.ReadFile(lerdExe)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dst, b, 0o755); err != nil {
			return err
		}
	}
	return nil
}

// sameProgram reports twin being lerd.exe: the same file when linked, the same
// bytes when copied.
func sameProgram(twin, lerdExe string) bool {
	a, errA := os.Stat(twin)
	b, errB := os.Stat(lerdExe)
	if errA != nil || errB != nil {
		return false
	}
	if os.SameFile(a, b) {
		return true
	}
	if a.Size() != b.Size() {
		return false
	}
	x, errA := os.ReadFile(twin)
	y, errB := os.ReadFile(lerdExe)
	return errA == nil && errB == nil && bytes.Equal(x, y)
}
