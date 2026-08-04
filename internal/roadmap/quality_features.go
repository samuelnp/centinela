package roadmap

import (
	"fmt"
	"strings"
)

// validateQualityFeatures checks each report entry against the roadmap: the
// feature exists, every score is in 1-10, the summary is non-empty, and no
// roadmap feature is left ungraded. There is deliberately NO minimum-score
// check — a self-assigned number gates nothing.
func validateQualityFeatures(r *Roadmap, features []QualityFeature) error {
	names := roadmapFeatureSet(r)
	seen := map[string]bool{}
	for _, f := range features {
		if !names[f.Name] {
			return fmt.Errorf("quality references unknown feature: %s", f.Name)
		}
		if err := validateScoreRange(f.Scores); err != nil {
			return fmt.Errorf("feature %s has invalid scores: %w", f.Name, err)
		}
		if strings.TrimSpace(f.Summary) == "" {
			return fmt.Errorf("feature %s summary is required", f.Name)
		}
		seen[f.Name] = true
	}
	for name := range names {
		if !seen[name] {
			return fmt.Errorf("quality missing feature: %s", name)
		}
	}
	return nil
}

// validateScoreRange names the offending field AND its value. This message is
// reserved for a genuinely out-of-range integer: every structural fault is
// caught earlier, in decodeQualityFeatures.
func validateScoreRange(s QualityScores) error {
	for _, f := range scoreFields(&s) {
		if *f.v < 1 || *f.v > 10 {
			return fmt.Errorf("score %q is %d, must be between 1 and 10", f.name, *f.v)
		}
	}
	return nil
}
