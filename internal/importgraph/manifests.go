package importgraph

import (
	"os"
	"path/filepath"
)

// detectKind returns the provider kind for a project by manifest presence,
// walking up from root to the nearest ancestor that carries a recognized
// manifest (mirroring how Go/Node/Python resolve a project root from a
// subdirectory). Precedence at each level is Go, then Node, then Python.
// Returns "" when no known manifest is found in root or any ancestor (→ Select
// yields ErrNoProvider → gate self-skips). The custom-script provider is never
// auto-selected; it must be requested explicitly.
func detectKind(root string) string {
	dir, err := filepath.Abs(root)
	if err != nil {
		dir = root
	}
	for {
		if k := kindAt(dir); k != "" {
			return k
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func kindAt(dir string) string {
	switch {
	case hasFile(dir, "go.mod"):
		return "go"
	case hasFile(dir, "package.json"):
		return "node"
	case hasFile(dir, "pyproject.toml"), hasFile(dir, "requirements.txt"),
		hasFile(dir, "setup.py"):
		return "python"
	default:
		return ""
	}
}

func hasFile(dir, name string) bool {
	info, err := os.Stat(filepath.Join(dir, name))
	return err == nil && !info.IsDir()
}
