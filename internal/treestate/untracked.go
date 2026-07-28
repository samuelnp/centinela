package treestate

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// UntrackedPaths returns the `??` entries of a porcelain=v1 status that lie
// outside .workflow/. Their status LINES never change when their contents do,
// so the caller must hash them separately.
func UntrackedPaths(status string) []string {
	paths := []string{}
	for _, ln := range strings.Split(status, "\n") {
		if !strings.HasPrefix(ln, "?? ") {
			continue
		}
		path := strings.Trim(strings.TrimSpace(ln[3:]), `"`)
		if path != "" && !strings.HasPrefix(path, excluded) {
			paths = append(paths, path)
		}
	}
	return paths
}

// HashUntracked returns one stable "path\x00hash" entry per path, sorted.
// A path that cannot be read contributes an error marker rather than being
// skipped: becoming unreadable is itself a change to the verified tree.
func HashUntracked(root string, paths []string) []string {
	entries := make([]string, 0, len(paths))
	for _, path := range paths {
		entries = append(entries, path+"\x00"+contentHash(filepath.Join(root, path)))
	}
	sort.Strings(entries)
	return entries
}

func contentHash(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "unreadable:" + err.Error()
	}
	return fmt.Sprintf("%x", sha256.Sum256(data))
}
