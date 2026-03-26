package setup

import (
	"os"
	"path/filepath"
)

// HookCmd is a single hook action entry in Claude's settings format.
type HookCmd struct {
	Type          string `json:"type"`
	Command       string `json:"command"`
	StatusMessage string `json:"statusMessage,omitempty"`
}

// HookGroup is a set of hooks, optionally scoped to a tool matcher.
type HookGroup struct {
	Matcher string    `json:"matcher,omitempty"`
	Hooks   []HookCmd `json:"hooks"`
}

// InjectHooks merges centinela hooks into the Claude settings file at path.
// It uses raw JSON maps at both levels so all existing keys are preserved.
// Returns true if the file was actually modified.
func InjectHooks(path string) (bool, error) {
	changed, data, err := buildHookSettings(path)
	if err != nil {
		return false, err
	}
	if !changed {
		return false, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return false, err
	}
	return true, os.WriteFile(path, data, 0644)
}
