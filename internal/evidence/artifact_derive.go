package evidence

import (
	"path/filepath"
	"sort"
	"strings"
)

// analyzedSpecsList renders the gatekeeper "Analyzed Specs" bullet list by
// globbing specs/*.feature (filesystem I/O). Returns a sorted "- specs/<name>"
// list, or a single FILL slot row when no spec files exist.
func analyzedSpecsList() string {
	files, _ := filepath.Glob("specs/*.feature")
	if len(files) == 0 {
		return "- " + FillSlot("list each .feature reviewed")
	}
	names := make([]string, 0, len(files))
	for _, f := range files {
		names = append(names, "- "+filepath.ToSlash(f))
	}
	sort.Strings(names)
	return strings.Join(names, "\n")
}
