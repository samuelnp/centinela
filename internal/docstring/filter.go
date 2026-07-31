package docstring

import (
	"path/filepath"
	"strings"
)

// InScope reports whether path is a file this package inspects: a non-test
// `.go` file under one of the configured roots, honoring IncludeInternal.
// Test files are excluded because test functions are not API surface — note
// that the G1 file-size rule *does* apply to them, so the difference here is
// deliberate rather than an oversight.
func InScope(path string, opts Options) bool {
	p := filepath.ToSlash(strings.TrimSpace(path))
	if !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
		return false
	}
	if !opts.IncludeInternal && hasSegment(p, "internal") {
		return false
	}
	return underRoots(p, roots(opts))
}

// underRoots reports whether p sits under any of the given root directories.
// A root of "." or "" means the whole tree.
func underRoots(p string, rs []string) bool {
	for _, r := range rs {
		r = strings.Trim(filepath.ToSlash(strings.TrimSpace(r)), "/")
		if r == "" || r == "." {
			return true
		}
		if p == r || strings.HasPrefix(p, r+"/") {
			return true
		}
	}
	return false
}

// hasSegment reports whether any slash-separated segment of p equals seg.
func hasSegment(p, seg string) bool {
	for _, s := range strings.Split(p, "/") {
		if s == seg {
			return true
		}
	}
	return false
}

// selected filters and sorts the caller's file list down to the in-scope set.
func selected(files []string, opts Options) []string {
	out := make([]string, 0, len(files))
	seen := map[string]bool{}
	for _, f := range files {
		p := filepath.ToSlash(strings.TrimSpace(f))
		if seen[p] || !InScope(p, opts) {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	return out
}
