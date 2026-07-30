// Acceptance: specs/token-diet.feature
package acceptance_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// Scenario: A changed roadmap is rendered again in the same session
func TestTD_ChangedRoadmapRenderedAgainSameSession(t *testing.T) {
	bin := tdBuildBin(t)
	dir := tdRepo(t)
	tdWriteRoadmap(t, dir, tdRoadmap1)
	staleDigest := tdCurrentDigest(t, dir)
	tdWriteRoadmap(t, dir, tdRoadmap2)
	tdWriteDigestState(t, dir, "s-1", staleDigest)

	out, code := tdHookContext(t, bin, dir, tdSessionPayload("s-1"))
	if code != 0 {
		t.Fatalf("hook context must exit 0: %s", out)
	}
	mustContain(t, out, tdRoadmapLine)
	newDigest := tdCurrentDigest(t, dir)
	t.Chdir(dir)
	got := roadmap.LoadSummaryState(roadmap.SummaryStatePath())
	if got.Digest != newDigest {
		t.Fatalf("digest state must update to current digest %s, got %+v", newDigest, got)
	}
}

// Scenario: A new session re-renders an unchanged roadmap
func TestTD_NewSessionReRendersUnchangedRoadmap(t *testing.T) {
	bin := tdBuildBin(t)
	dir := tdRepo(t)
	tdWriteRoadmap(t, dir, tdRoadmap1)
	digest := tdCurrentDigest(t, dir)
	tdWriteDigestState(t, dir, "s-1", digest)

	out, code := tdHookContext(t, bin, dir, tdSessionPayload("s-2"))
	if code != 0 {
		t.Fatalf("hook context must exit 0: %s", out)
	}
	mustContain(t, out, tdRoadmapLine)
}

// Scenario: Roadmap reformatting alone does not force a re-render
func TestTD_RoadmapReformattingAloneDoesNotForceReRender(t *testing.T) {
	bin := tdBuildBin(t)
	dir := tdRepo(t)
	tdWriteRoadmap(t, dir, tdRoadmap1)
	digest := tdCurrentDigest(t, dir)
	tdWriteDigestState(t, dir, "s-1", digest)
	tdWriteRoadmap(t, dir, tdRoadmap1Reformatted)

	out, code := tdHookContext(t, bin, dir, tdSessionPayload("s-1"))
	if code != 0 {
		t.Fatalf("hook context must exit 0: %s", out)
	}
	mustNotContain(t, out, tdRoadmapLine)
}
