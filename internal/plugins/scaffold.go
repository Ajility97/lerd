package plugins

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// skeleton is the boilerplate plugin `lerd plugin new` copies. It reads its own
// name from LERD_PLUGIN_NAME everywhere, so only plugin.yaml needs renaming.
//
//go:embed all:skeleton
var skeleton embed.FS

// Scaffold writes a new plugin called name into dest, which must not exist.
func Scaffold(name, dest string) error {
	if !nameRe.MatchString(name) {
		return fmt.Errorf("invalid plugin name %q: use lowercase letters, digits and dashes", name)
	}
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("%s already exists", dest)
	}
	return fs.WalkDir(skeleton, "skeleton", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(dest, strings.TrimPrefix(path, "skeleton"))
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := skeleton.ReadFile(path)
		if err != nil {
			return err
		}
		if path == "skeleton/"+ManifestFile {
			data = []byte(strings.Replace(string(data), "\nname: hello\n", "\nname: "+name+"\n", 1))
		}
		// embed drops file modes, so anything under bin/ is the executable part.
		mode := os.FileMode(0o644)
		if strings.HasPrefix(path, "skeleton/bin/") {
			mode = 0o755
		}
		return os.WriteFile(target, data, mode)
	})
}
