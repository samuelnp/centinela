package setup

import (
	"encoding/json"
	"os"
)

const opencodeSchema = "https://opencode.ai/config.json"

// InjectOpenCodeConfig merges Centinela defaults into opencode.json.
// Existing unrelated keys are preserved.
func InjectOpenCodeConfig(path string) (bool, error) {
	changed, data, err := buildOpenCodeConfig(path)
	if err != nil {
		return false, err
	}
	if !changed {
		return false, nil
	}
	return true, os.WriteFile(path, data, 0644)
}

func mergeInstructions(raw map[string]json.RawMessage) bool {
	var values []string
	_ = json.Unmarshal(raw["instructions"], &values)
	ordered := []string{}
	for _, v := range values {
		if v == "AGENTS.md" || v == "CLAUDE.md" || hasValue(ordered, v) {
			continue
		}
		ordered = append(ordered, v)
	}
	ordered = append(ordered, "AGENTS.md", "CLAUDE.md")
	if sameStrings(values, ordered) {
		return false
	}
	raw["instructions"], _ = json.Marshal(ordered)
	return true
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func hasValue(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}

func lookup(raw map[string]json.RawMessage, key string) string {
	var v string
	_ = json.Unmarshal(raw[key], &v)
	return v
}
