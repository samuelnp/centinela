package workflow

import (
	"fmt"
	"os"

	"github.com/samuelnp/centinela/internal/gatereport"
	"github.com/samuelnp/centinela/internal/treestate"
)

const reverify = "Spawn a FRESH verifier; the previous verdict cannot certify a changed tree."

// VerificationFresh reports whether the verifier's stamp still describes the
// current tree. Called from the `complete` path ONLY (D6) — it shells out to
// git, which ValidateArtifacts must never do. Legacy workflows are exempt.
func VerificationFresh(feature, root string, run treestate.Runner) error {
	if !featureUsesAdversarialVerifier(feature) {
		return nil
	}
	report := GatekeeperReportPath(feature)
	data, err := os.ReadFile(report)
	if err != nil {
		return fmt.Errorf("gatekeeper report not found: %s", report)
	}
	// Admissibility failures — absent/malformed block, no commands record, an
	// unstamped revision — belong to gatereport.Assess, which names the remedy.
	// Reporting staleness for them would mask the message the operator needs
	// (an `artifact new` stub would read as "stale", not "ungrounded").
	if gatereport.Assess(string(data)) != nil {
		return nil
	}
	recorded, err := gatereport.ParseVerification(string(data))
	if err != nil {
		return nil
	}
	current, err := treestate.Stamp(root, run)
	if err != nil {
		return fmt.Errorf("cannot determine current tree state to check %s: %w", report, err)
	}
	return compareStamp(recorded, current, root, run)
}

// compareStamp names which half of the binding drifted: HEAD moving means
// fixes were committed on top of the verified commit, an equal HEAD with a
// different digest means fixes were made in place (D3 — the case a HEAD-only
// comparison would wrongly admit).
//
// A moved HEAD is forgiven ONLY when every path in the range is roadmap state
// (D5): a roadmap mutation commits itself now, so recording a deferral must not
// void the verdict the same verifier is about to stamp. Any other path in the
// range still stales it, with today's message.
func compareStamp(recorded gatereport.Verification, current treestate.Snapshot, root string, run treestate.Runner) error {
	if recorded.Revision != current.Revision &&
		!revisionRangeIsRoadmapStateOnly(root, recorded.Revision, run) {
		return fmt.Errorf("gatekeeper verification is stale (verified %s, HEAD is now %s). %s",
			short(recorded.Revision), short(current.Revision), reverify)
	}
	if recorded.TreeDigest != current.Digest {
		return fmt.Errorf("gatekeeper verification is stale (verified tree %s, tree is now %s — uncommitted changes since verification). %s",
			short(recorded.TreeDigest), short(current.Digest), reverify)
	}
	return nil
}

// short renders a sha-ish value at review length, keeping any algorithm prefix.
func short(value string) string {
	if len(value) <= 14 {
		return value
	}
	return value[:14]
}
