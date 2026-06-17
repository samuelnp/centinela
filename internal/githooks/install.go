// Package githooks installs and removes the centinela pre-commit hook via a
// marker-delimited splice into hooksDir/pre-commit, never clobbering user
// content. It is stdlib-only (a leaf): it imports no project package.
package githooks

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	BeginMarker = "# >>> centinela >>>"
	EndMarker   = "# <<< centinela <<<"
)

// Block is the managed hook body, fenced by the markers.
const Block = BeginMarker + "\n" +
	"#!/bin/sh\n" +
	"# Managed by centinela — do not edit between the markers.\n" +
	"centinela precommit\n" +
	EndMarker + "\n"

// Install writes/refreshes the centinela block in hooksDir/pre-commit,
// preserving content outside the markers, creating the dir+file when absent,
// and (re)asserting the executable bit. Idempotent: re-install of an unchanged
// block returns changed=false.
func Install(hooksDir string) (changed bool, err error) {
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return false, err
	}
	path := filepath.Join(hooksDir, "pre-commit")
	existing := readHook(path)
	next, changed := splice(existing, Block)
	if err := os.WriteFile(path, []byte(next), 0o755); err != nil {
		return false, err
	}
	if err := os.Chmod(path, 0o755); err != nil {
		return false, err
	}
	return changed, nil
}

// Uninstall removes the centinela block from hooksDir/pre-commit, leaving user
// content intact. If the file is left empty or only a bare shebang, it is
// deleted. A missing file or absent block is a no-op (changed=false).
func Uninstall(hooksDir string) (changed bool, err error) {
	path := filepath.Join(hooksDir, "pre-commit")
	existing, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	next, changed := removeBlock(string(existing))
	if !changed {
		return false, nil
	}
	if isEmptyHook(next) {
		if err := os.Remove(path); err != nil {
			return false, err
		}
		return true, nil
	}
	if err := os.WriteFile(path, []byte(next), 0o755); err != nil {
		return false, err
	}
	return true, nil
}

func readHook(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

// isEmptyHook reports whether what remains is blank or just a bare shebang.
func isEmptyHook(s string) bool {
	for _, line := range strings.Split(s, "\n") {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "#!") {
			continue
		}
		return false
	}
	return true
}
