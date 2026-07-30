package roadmap

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// SummaryStateFile holds the per-worktree render state for the one-line roadmap
// summary injected on every user prompt. It is git-ignored runtime state, not
// durable project state.
const SummaryStateFile = ".workflow/.roadmap-digest"

// SummaryState records which session last saw which roadmap projection.
type SummaryState struct {
	SessionID string `json:"sessionId"`
	Digest    string `json:"digest"`
}

// SummaryStatePath returns the digest state path, relative to the worktree root
// so two worktrees of the same repo keep independent state.
func SummaryStatePath() string { return SummaryStateFile }

// SummaryDigest hashes a STABLE PROJECTION of the summary-relevant state — the
// planned/inProgress/done counts plus each phase's name and each feature's
// name+status in declaration order — never the raw file bytes. Reformatting
// roadmap.json (whitespace, key order) therefore does not force a re-render,
// while a real status or membership change does.
func SummaryDigest(r *Roadmap) string {
	if r == nil {
		return ""
	}
	planned, inProgress, done := r.Summary()
	var b strings.Builder
	fmt.Fprintf(&b, "counts:%d/%d/%d\n", planned, inProgress, done)
	for _, phase := range r.Phases {
		b.WriteString("phase:" + phase.Name + "\n")
		for _, f := range phase.Features {
			b.WriteString("feature:" + f.Name + "=" + FeatureStatus(f.Name) + "\n")
		}
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])[:16]
}

// ShouldRenderSummary fails OPEN: an absent session signal always renders, as
// does a new session or a changed digest.
func ShouldRenderSummary(prev SummaryState, sessionID, digest string) bool {
	if sessionID == "" {
		return true
	}
	return prev.SessionID != sessionID || prev.Digest != digest
}
