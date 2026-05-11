// Package gitdiff resolves the set of files changed on the current branch
// relative to a configurable base ref. The set is consumed by file-scoped
// gates (G1, G11) when centinela validate runs in diff-aware mode.
package gitdiff

import "path/filepath"

// Set is an immutable lookup of changed file paths, normalized to
// forward-slash relative paths.
type Set struct {
	paths map[string]struct{}
}

// NewSet builds a Set from a slice of paths. Empty strings are skipped.
// Paths are normalized to forward slashes for cross-platform matching.
func NewSet(paths []string) *Set {
	m := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		if p == "" {
			continue
		}
		m[filepath.ToSlash(p)] = struct{}{}
	}
	return &Set{paths: m}
}

// Contains reports whether the given path (any separator) is in the set.
func (s *Set) Contains(path string) bool {
	if s == nil {
		return false
	}
	_, ok := s.paths[filepath.ToSlash(path)]
	return ok
}

// Len returns the number of paths in the set.
func (s *Set) Len() int {
	if s == nil {
		return 0
	}
	return len(s.paths)
}

// HasPrefix reports whether any path in the set begins with the given
// prefix (matched on slash-normalized paths). Used by G11 to test
// whether any locale file changed without enumerating the dir.
func (s *Set) HasPrefix(prefix string) bool {
	if s == nil {
		return false
	}
	p := filepath.ToSlash(prefix)
	for k := range s.paths {
		if len(k) >= len(p) && k[:len(p)] == p {
			return true
		}
	}
	return false
}
