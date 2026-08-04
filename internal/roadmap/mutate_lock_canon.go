package roadmap

import "path/filepath"

// canonicalPath resolves path to one spelling per FILE, symlinks included.
//
// filepath.Abs alone is not enough: it never resolves symlinks, so the same
// roadmap.json reached as /real/proj/... and as /link/proj/... hashed to two
// different lock names, two lock files were created side by side in the same
// .git, and the mutation race was silently back — with a green "✓ Committed"
// printed for a record that had been overwritten. macOS symlinks /tmp and /var,
// so this is an everyday path spelling, not an exotic one.
//
// The file itself is resolved when it exists; otherwise its DIRECTORY is, so a
// not-yet-created roadmap.json still canonicalizes. Every failure falls back to
// the absolute path — a lock under an unresolvable name is still a lock, and
// nothing here may panic or quietly skip locking.
func canonicalPath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = filepath.Clean(path)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return resolved
	}
	dir, base := filepath.Split(abs)
	resolved, err := filepath.EvalSymlinks(filepath.Clean(dir))
	if err != nil {
		return abs
	}
	return filepath.Join(resolved, base)
}

// canonicalDir resolves an existing directory's symlinks so the lock directory
// itself has one spelling. Two processes that resolved the repo root
// differently must land on the same lock PATH, not merely the same inode.
func canonicalDir(dir string) string {
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		return resolved
	}
	return dir
}
