package roadmap

import (
	"encoding/json"
	"fmt"
	"os"
)

// appendPromotionArtifacts records the promoted finding in the analysis +
// quality JSON (raw-preserving) and appends provenance bullets to their .md
// companions. Each file write is atomic (temp-file+rename).
func appendPromotionArtifacts(slug, summary string, scores QualityScores, f *BacklogFinding) error {
	return appendScoreArtifacts(slug, summary, scores, provenanceBullet(slug, f))
}

// appendFeatureEntry appends entry to the top-level "features" array of a JSON
// artifact, preserving every other field and untouched entry verbatim.
func appendFeatureEntry(path string, entry json.RawMessage) error {
	data, err := os.ReadFile(path)
	if err != nil {
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
	feats = append(feats, entry)
	encoded, err := compactBytes(feats)
	if err != nil {
		return err
	}
	top["features"] = encoded
	return writeArtifact(path, top)
}

// provenanceBullet records the original source + deferredAt of a promoted
// finding so the provenance survives even though it is stripped from the
// roadmap feature entry.
func provenanceBullet(slug string, f *BacklogFinding) string {
	src := "unknown"
	if f.Source != nil {
		src = f.Source.Feature
		if f.Source.Role != "" {
			src += "/" + f.Source.Role
		}
	}
	return fmt.Sprintf("- Promoted from Backlog: %s (source=%s, deferredAt=%s)",
		slug, src, f.DeferredAt)
}
