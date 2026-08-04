package roadmap

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
)

// lockName is the stem of the advisory lock file. It is keyed by a digest of
// roadmap.json's CANONICAL path so two checkouts (or two worktrees) never share
// a lock, and two processes on the same file always agree on one.
const lockName = "centinela-roadmap-"

// stateLockPath returns where to put the advisory lock for roadmap.json.
//
// Deliberately NOT a sibling `.workflow/roadmap.json.lock`: Centinela does not
// manage a consumer project's .gitignore, so a sibling would show up as an
// untracked file in every project that ever ran a mutation — dirtying the tree
// is the precise failure this feature exists to remove, and an aborted ff-merge
// on a dirty tree is what started it.
//
// Preference order:
//  1. the git directory, which is per-checkout, shared by every user of that
//     checkout, and never reported by `git status`;
//  2. the OS temp dir, for a roadmap.json that is not in a repository at all
//     (the supported "no git" mutation path).
func stateLockPath(path string) string {
	canon := canonicalPath(path)
	sum := sha256.Sum256([]byte(canon))
	name := lockName + hex.EncodeToString(sum[:8]) + ".lock"
	if dir, ok := gitDirFor(canon); ok {
		return filepath.Join(dir, name)
	}
	return filepath.Join(os.TempDir(), name)
}

// gitDirFor walks up from roadmap.json looking for `.git`, resolving the
// `gitdir:` pointer a linked worktree uses. Reading the pointer directly keeps
// the domain free of git subprocesses — running git is the CLI's job, behind
// the Committer seam.
func gitDirFor(abs string) (string, bool) {
	for dir := filepath.Dir(abs); ; {
		candidate := filepath.Join(dir, ".git")
		if info, err := os.Stat(candidate); err == nil {
			if info.IsDir() {
				return canonicalDir(candidate), true
			}
			if resolved, ok := readGitDirPointer(candidate, dir); ok {
				return canonicalDir(resolved), true
			}
			return "", false
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// readGitDirPointer parses the one-line `gitdir: <path>` file a linked worktree
// carries in place of a .git directory. A relative target resolves against the
// worktree root, matching git's own rule.
func readGitDirPointer(file, root string) (string, bool) {
	data, err := os.ReadFile(file)
	if err != nil {
		return "", false
	}
	target := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(string(data)), "gitdir:"))
	if target == "" {
		return "", false
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(root, target)
	}
	if info, err := os.Stat(target); err != nil || !info.IsDir() {
		return "", false
	}
	return target, true
}
