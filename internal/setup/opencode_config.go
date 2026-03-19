package setup

import (
	"encoding/json"
	"os"
)

const opencodeSchema = "https://opencode.ai/config.json"

// InjectOpenCodeConfig merges Centinela defaults into opencode.json.
// Existing unrelated keys are preserved.
func InjectOpenCodeConfig(path string) (bool, error) {
	raw := map[string]json.RawMessage{}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &raw)
	}

	changed := false
	if lookup(raw, "$schema") == "" {
		raw["$schema"], _ = json.Marshal(opencodeSchema)
		changed = true
	}

	if mergeInstructions(raw) {
		changed = true
	}
	if !changed {
		return false, nil
	}

	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return false, err
	}
	return true, os.WriteFile(path, data, 0644)
}

func mergeInstructions(raw map[string]json.RawMessage) bool {
	var values []string
	_ = json.Unmarshal(raw["instructions"], &values)
	if hasValue(values, "CLAUDE.md") {
		return false
	}
	values = append(values, "CLAUDE.md")
	raw["instructions"], _ = json.Marshal(values)
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
