package roadmap

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// crudPromoteDir chdirs into a temp dir seeded with roadmap body plus empty
// analysis/quality artifacts, so a promote can append and re-validate.
func crudPromoteDir(t *testing.T, body string) {
	t.Helper()
	crudChdir(t, body)
	os.WriteFile(RoadmapAnalysisFile, []byte(`{"role":"senior-product-manager","features":[]}`), 0o644)                 //nolint:errcheck
	os.WriteFile(RoadmapQualityFile, []byte(`{"role":"roadmap-quality-evaluator","threshold":9,"features":[]}`), 0o644) //nolint:errcheck
	os.WriteFile(RoadmapAnalysisMarkdown, []byte("# analysis\n"), 0o644)                                                //nolint:errcheck
	os.WriteFile(RoadmapQualityMarkdown, []byte("# quality\n"), 0o644)                                                  //nolint:errcheck
}

const draftBody = `{"phases":[{"name":"Phase 1: Foundations","features":[` +
	`{"name":"new-widget","draft":true}]}]}`

// TestPromoteDraftInPlace_Finalizes clears the draft without moving and writes artifacts.
func TestPromoteDraftInPlace_Finalizes(t *testing.T) {
	crudPromoteDir(t, draftBody)
	scores, _ := ParseScores("9,9,9,9,9,9")
	if _, err := Promote(RoadmapFile, PromoteRequest{Slug: "new-widget", Scores: scores}); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	road := string(crudBytes(t, RoadmapFile))
	if !strings.Contains(road, "Phase 1: Foundations") || !strings.Contains(road, "new-widget") {
		t.Fatal("feature must remain in place (no move)")
	}
	if strings.Contains(road, `"draft":true`) || strings.Contains(road, `"draft": true`) {
		t.Fatal("draft flag must be cleared")
	}
	if !strings.Contains(string(crudBytes(t, RoadmapAnalysisFile)), "new-widget") {
		t.Fatal("analysis must gain new-widget")
	}
	q := string(crudBytes(t, RoadmapQualityFile))
	if !strings.Contains(q, "new-widget") || !strings.Contains(q, `"overall":9`) {
		t.Fatalf("quality must gain new-widget with overall 9: %s", q)
	}
	r, err := Load()
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if err := ValidateAnalysis(r); err != nil {
		t.Fatalf("ValidateAnalysis: %v", err)
	}
	if err := ValidateQuality(r); err != nil {
		t.Fatalf("ValidateQuality: %v", err)
	}
}

// TestPromote_NonDraftNonBacklogError errors and writes nothing.
func TestPromote_NonDraftNonBacklogError(t *testing.T) {
	body := `{"phases":[{"name":"Phase 1: Foundations","features":[{"name":"auth-service"}]}]}`
	crudPromoteDir(t, body)
	before := crudBytes(t, RoadmapFile)
	scores, _ := ParseScores("9,9,9,9,9,9")
	_, err := Promote(RoadmapFile, PromoteRequest{Slug: "auth-service", Scores: scores})
	if err == nil || !strings.Contains(err.Error(), "auth-service") {
		t.Fatalf("expected neither-draft-nor-backlog error, got %v", err)
	}
	if !bytes.Equal(before, crudBytes(t, RoadmapFile)) {
		t.Fatal("rejected promote must be byte-identical")
	}
}

// TestDraftSummaryFallback exercises the summary → description → slug fallback.
func TestDraftSummaryFallback(t *testing.T) {
	if got := draftSummary(PromoteRequest{Summary: " x "}, Feature{}); got != "x" {
		t.Fatalf("explicit summary wins: %q", got)
	}
	if got := draftSummary(PromoteRequest{}, Feature{Description: "d"}); got != "d" {
		t.Fatalf("description fallback: %q", got)
	}
	if got := draftSummary(PromoteRequest{Slug: "s"}, Feature{}); got != "s" {
		t.Fatalf("slug fallback: %q", got)
	}
	if !strings.Contains(draftFinalizeBullet("s"), "s") {
		t.Fatal("bullet must name the slug")
	}
}
