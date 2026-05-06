package setup

import (
	"encoding/json"
	"os"
)

func buildOpenCodeConfig(path string) (bool, []byte, error) {
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
	if mergeOpenCodeAgents(raw) {
		changed = true
	}
	if !changed {
		return false, nil, nil
	}
	data, err := json.MarshalIndent(raw, "", "  ")
	return true, data, err
}
