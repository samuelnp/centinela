package scaffold

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed all:assets
var assets embed.FS

// Result summarises what Extract created and what it skipped.
type Result struct {
	Created []string
	Skipped []string
}

// Extract copies embedded assets into dir.
// Files that already exist are skipped — safe to run multiple times.
func Extract(dir string) (Result, error) {
	var result Result
	err := fs.WalkDir(assets, "assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel := path[len("assets/"):]
		target := filepath.Join(dir, rel)

		if _, err := os.Stat(target); err == nil {
			result.Skipped = append(result.Skipped, rel)
			return nil
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		data, err := assets.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.WriteFile(target, data, 0644); err != nil {
			return err
		}
		result.Created = append(result.Created, rel)
		return nil
	})
	return result, err
}

// ReadAsset reads a scaffold asset by project-relative path.
func ReadAsset(path string) ([]byte, error) {
	return assets.ReadFile(filepath.Join("assets", path))
}

// ListAssetFiles returns scaffold asset file paths matching prefix/suffix.
func ListAssetFiles(prefix, suffix string) ([]string, error) {
	var out []string
	err := fs.WalkDir(assets, "assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel := strings.TrimPrefix(path, "assets/")
		if strings.HasPrefix(rel, prefix) && strings.HasSuffix(rel, suffix) {
			out = append(out, rel)
		}
		return nil
	})
	return out, err
}
