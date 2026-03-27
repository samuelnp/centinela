package setup

import (
	"encoding/json"
	"os"
)

func buildHookSettings(path string) (bool, []byte, error) {
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
		if !ensureStatusLine(rawSettings) {
			return false, nil, nil
		}
	} else {
		ensureStatusLine(rawSettings)
	}
	rawHooks["PreToolUse"], _ = json.Marshal(pre)
	rawHooks["PostToolUse"], _ = json.Marshal(post)
	rawHooks["UserPromptSubmit"], _ = json.Marshal(prompt)
	rawSettings["hooks"], _ = json.Marshal(rawHooks)
	data, err := json.MarshalIndent(rawSettings, "", "  ")
	return true, data, err
}
