package roadmap

import (
	"encoding/json"
	"fmt"
	"os"
)

// preflightArtifacts confirms both artifact JSON files exist and parse and both
// .md companions exist BEFORE roadmap.json is written, so a missing/corrupt
// artifact aborts promote with nothing written (no half-promoted state).
func preflightArtifacts() error {
	for _, p := range []string{RoadmapAnalysisFile, RoadmapQualityFile} {
		if err := checkArtifactJSON(p); err != nil {
			return err
		}
	}
	for _, p := range []string{RoadmapAnalysisMarkdown, RoadmapQualityMarkdown} {
		if _, err := os.Stat(p); err != nil {
			return fmt.Errorf("roadmap artifact markdown missing: %s", p)
		}
	}
	return nil
}

// checkArtifactJSON confirms an artifact file exists, is a JSON object, and has
// a well-formed (if present) "features" array, without mutating it.
func checkArtifactJSON(p string) error {
	data, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("roadmap artifact json missing: %s", p)
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return fmt.Errorf("invalid artifact json %s: %w", p, err)
	}
	if raw, ok := top["features"]; ok {
		var feats []json.RawMessage
		if err := json.Unmarshal(raw, &feats); err != nil {
			return fmt.Errorf("invalid features array in %s: %w", p, err)
		}
	}
	return nil
}
