package roadmap

import (
	"encoding/json"
	"fmt"
	"os"
)

// removeFeatureEntries drops every feature entry whose name is in remove from the
// top-level "features" array of a JSON artifact (analysis/quality), preserving
// every other field and untouched entry verbatim, then writes it atomically. A
// missing artifact file is a no-op (nothing to prune). Mirrors appendFeatureEntry.
func removeFeatureEntries(path string, remove map[string]bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return fmt.Errorf("invalid artifact json %s: %w", path, err)
	}
	var feats []json.RawMessage
	if raw, ok := top["features"]; ok {
		if err := json.Unmarshal(raw, &feats); err != nil {
			return fmt.Errorf("invalid features array in %s: %w", path, err)
		}
	}
	kept := feats[:0:0]
	for _, f := range feats {
		if name, _ := featureName(f); !remove[name] {
			kept = append(kept, f)
		}
	}
	encoded, err := compactBytes(kept)
	if err != nil {
		return err
	}
	top["features"] = encoded
	return writeArtifact(path, top)
}
