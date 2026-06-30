package analyze

import (
	"os"
	"path/filepath"
)

// sourceDirs are the root-level directories whose presence (when non-empty)
// signals an existing codebase. Kept deliberately small: these are the
// conventional source roots across the ecosystems Centinela governs. An empty
// directory of any of these names is NOT a signal (a freshly scaffolded repo
// may carry an empty src/ — see the spec's empty-src scenario).
var sourceDirs = []string{"src", "app", "lib", "cmd", "pkg", "internal"}

// HasSource reports whether root looks like an existing (brownfield) codebase
// using only a cheap, root-only inspection — no recursion, no file reads. It
// runs on every UserPromptSubmit hook, so cost must stay O(entries at root).
//
// A repo is brownfield if EITHER:
//   - the root contains any known manifest (the same set manifests.go knows,
//     sourced from manifestTable so the two never drift), OR
//   - the root contains a non-empty conventional source directory.
//
// It deliberately does not descend into source dirs: deeply nested source with
// no root-level signal reads as greenfield (the detector is intentionally
// shallow to stay cheap).
func HasSource(root string) bool {
	// Strong signal: a known manifest at the root. Reuse manifestTable keys so
	// the detector's manifest set is identical to the analyzer's by construction.
	for name := range manifestTable {
		if _, err := os.Stat(filepath.Join(root, name)); err == nil {
			return true
		}
	}
	// Weaker signal: a conventional source dir that actually holds something.
	for _, dir := range sourceDirs {
		if dirHasEntry(filepath.Join(root, dir)) {
			return true
		}
	}
	return false
}

// dirHasEntry reports whether path is a directory containing at least one
// entry. It reads just a single entry (ReadDir(1)) so an enormous source tree
// costs no more than a tiny one. A non-existent path or a plain file is not a
// signal.
func dirHasEntry(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false
	}
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	names, _ := f.Readdirnames(1)
	return len(names) > 0
}
