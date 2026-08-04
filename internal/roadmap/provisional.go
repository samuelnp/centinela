package roadmap

import (
	"encoding/json"
	"fmt"
	"os"
)

// ProvisionalKey is the JSON field that marks a grading artifact as SEEDED by
// tooling rather than produced by an evaluator pass. writeArtifact preserves
// every non-"features" key, so the mark survives the appends `roadmap promote`
// makes and can only be cleared deliberately, by a human or a real evaluator.
const ProvisionalKey = "provisional"

// ArtifactIsProvisional reports whether the grading artifact at path exists and
// carries the provisional mark. A missing or unreadable file is NOT provisional
// — callers that care about existence check that separately, and this must never
// turn "absent" into "seeded".
//
// Exported because the strict setup cascade needs the same answer ValidateAnalysis
// and ValidateQuality give: those check the mark, but the cascade's rungs tested
// file EXISTENCE only, so seeded artifacts satisfied `hook setup` on the very tree
// where `start` refused them. Two surfaces, one question, one implementation.
func ArtifactIsProvisional(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var top struct {
		Provisional bool `json:"provisional"`
	}
	if err := json.Unmarshal(data, &top); err != nil {
		return false
	}
	return top.Provisional
}

// provisionalRefusal explains why a seeded artifact does not satisfy a rung that
// exists to demand an evaluation. Naming the exact file and the exact edit keeps
// the refusal actionable instead of merely correct.
func provisionalRefusal(path, role string) error {
	return fmt.Errorf(
		"%s is provisional: it was seeded by `centinela roadmap promote` so that "+
			"work could start, and records only CLI-supplied scores — no %s pass has "+
			"run. Run one, then remove %q from %s",
		path, role, ProvisionalKey, path)
}

// A seeded artifact carries the evaluator's role string because promote must be
// able to append to it, and because the guided profile treats these artifacts as
// advisory anyway. That makes the role alone a forgeable claim: without this
// mark, one guided promote would permanently satisfy the STRICT greenfield
// grading rung — pin strict afterwards and `start` would sail through a cascade
// no evaluator ever touched. The mark closes that cross-profile hole without
// removing seeding, which is what keeps the guided cold start traversable.
