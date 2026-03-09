package setup

import (
	"encoding/json"
	"os"
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
	rawSettings := map[string]json.RawMessage{}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &rawSettings)
	}

	rawHooks := map[string]json.RawMessage{}
	if h, ok := rawSettings["hooks"]; ok {
		_ = json.Unmarshal(h, &rawHooks)
	}

	var pre, post, prompt []HookGroup
	_ = json.Unmarshal(rawHooks["PreToolUse"], &pre)
	_ = json.Unmarshal(rawHooks["PostToolUse"], &post)
	_ = json.Unmarshal(rawHooks["UserPromptSubmit"], &prompt)

	if !mergeHooks(&pre, &post, &prompt) {
		return false, nil
	}

	rawHooks["PreToolUse"], _ = json.Marshal(pre)
	rawHooks["PostToolUse"], _ = json.Marshal(post)
	rawHooks["UserPromptSubmit"], _ = json.Marshal(prompt)
	rawSettings["hooks"], _ = json.Marshal(rawHooks)

	if err := os.MkdirAll(".claude", 0755); err != nil {
		return false, err
	}
	data, err := json.MarshalIndent(rawSettings, "", "  ")
	if err != nil {
		return false, err
	}
	return true, os.WriteFile(path, data, 0644)
}
