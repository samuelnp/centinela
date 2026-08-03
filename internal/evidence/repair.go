package evidence

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

const (
	// placeholderFeature is the obvious unfilled slot the skeleton prints for
	// `feature` when no feature is known.
	placeholderFeature = "<feature-slug>"
	// unfilledHandoffSlot is its sibling for `handoffTo`. Non-empty, so it
	// clears the generic completeness check, but not a real role name — so
	// workflow.CheckHandoffTo refuses it with its actionable message if it is
	// ever pasted verbatim. "The author must decide" beats "confidently wrong".
	unfilledHandoffSlot = "<successor-role>"
)

// Repair removes orphaned `<feature>-<role>.json.tmp` files left behind by a
// crashed atomic write. It is idempotent and safe to re-run; the caller may
// invoke it before every `set`/`append` flow without risk. Glob errors
// surface only on malformed patterns — the literal pattern below cannot
// trip that path, so the error return is wrapped defensively only.
func Repair(feature string) ([]string, error) {
	prefix := feature + "-"
	matches, _ := filepath.Glob(filepath.Join(workflow.WorkflowDir, prefix+"*.json"+tempSuffix))
	removed := []string{}
	for _, path := range matches {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return removed, fmt.Errorf("evidence repair remove %s: %w", path, err)
		}
		removed = append(removed, path)
	}
	return removed, nil
}

// SchemaSkeleton returns the JSON skeleton for the given role rendered for
// embedding in prompts. The CLI version is stamped so the file shows what
// binary produced the skeleton.
//
// feature is the feature the caller resolved (see ResolveActiveFeature), or ""
// when none could be resolved. With a feature, handoffTo comes from the SAME
// derivation `evidence init` and the completion gate use — one derivation,
// three callers. Without one, both unknown fields print their placeholder
// rather than a legacy static guess the gate would then refuse.
func SchemaSkeleton(feature string, role Role, cliVersion string) ([]byte, error) {
	f := feature
	if f == "" {
		f = placeholderFeature
	}
	skel := Skeleton(f, role, cliVersion)
	// merge-steward is out-of-band: handoffForRole answers the fixed literal
	// "complete" for it independent of any feature, so overriding that already
	// correct answer with a placeholder would only break a valid stub.
	if feature == "" && role != orchestration.RoleMergeSteward {
		skel.HandoffTo = unfilledHandoffSlot
	}
	return skel.MarshalJSON()
}
