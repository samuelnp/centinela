// Acceptance: specs/token-diet.feature
package acceptance_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// tdCurrentDigest computes SummaryDigest for the roadmap already written at
// dir, by chdir-ing into it (Load() and SummaryStatePath() are both
// worktree-relative).
func tdCurrentDigest(t *testing.T, dir string) string {
	t.Helper()
	t.Chdir(dir)
	r, err := roadmap.Load()
	if err != nil {
		t.Fatalf("load roadmap: %v", err)
	}
	return roadmap.SummaryDigest(r)
}

func tdWriteDigestState(t *testing.T, dir, sessionID, digest string) {
	t.Helper()
	t.Chdir(dir)
	if err := roadmap.SaveSummaryState(roadmap.SummaryStatePath(), roadmap.SummaryState{
		SessionID: sessionID, Digest: digest,
	}); err != nil {
		t.Fatalf("save digest state: %v", err)
	}
}

// Scenario: The first hook context of a session renders the roadmap summary
func TestTD_FirstHookContextOfSessionRendersRoadmapSummary(t *testing.T) {
	bin := tdBuildBin(t)
	dir := tdRepo(t)
	tdWriteRoadmap(t, dir, tdRoadmap1)

	out, code := tdHookContext(t, bin, dir, tdSessionPayload("s-1"))
	if code != 0 {
		t.Fatalf("hook context must exit 0: %s", out)
	}
	mustContain(t, out, tdRoadmapLine)

	digest := tdCurrentDigest(t, dir)
	t.Chdir(dir)
	got := roadmap.LoadSummaryState(roadmap.SummaryStatePath())
	if got.SessionID != "s-1" || got.Digest != digest {
		t.Fatalf("digest state = %+v, want session s-1 digest %s", got, digest)
	}
}

// Scenario: An unchanged roadmap is not re-rendered in the same session
func TestTD_UnchangedRoadmapNotReRenderedSameSession(t *testing.T) {
	bin := tdBuildBin(t)
	dir := tdRepo(t)
	tdWriteRoadmap(t, dir, tdRoadmap1)
	tdWriteWorkflow(t, dir, "widget", "tests", "", false)
	digest := tdCurrentDigest(t, dir)
	tdWriteDigestState(t, dir, "s-1", digest)

	out, code := tdHookContext(t, bin, dir, tdSessionPayload("s-1"))
	if code != 0 {
		t.Fatalf("hook context must exit 0: %s", out)
	}
	mustNotContain(t, out, tdRoadmapLine)
	mustContain(t, out, "widget") // active-workflow panel unaffected
	mustContain(t, out, "widget-edge-cases.md")
}
