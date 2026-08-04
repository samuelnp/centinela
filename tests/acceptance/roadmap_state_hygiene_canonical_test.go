// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import (
	"encoding/json"
	"strings"
	"testing"
)

// Scenario: every phase renders one feature object per line on every write
func TestRsh_EveryPhaseRendersOneFeaturePerLine(t *testing.T) {
	bin := buildCent(t)
	dir := rshRepo(t, rshBaseRoadmap)
	if out, code := runCent(t, bin, dir, "roadmap", "defer", "canon-thing", "--summary", "x"); code != 0 {
		t.Fatalf("defer exit=%d\n%s", code, out)
	}
	body := mustRead(t, dir+"/.workflow/roadmap.json")
	for _, ln := range strings.Split(body, "\n") {
		s := strings.TrimSpace(strings.TrimSuffix(ln, ","))
		if !strings.HasPrefix(s, `{"name"`) {
			continue
		}
		if !json.Valid([]byte(s)) {
			t.Fatalf("expected exactly one compact feature object per line, got %q", s)
		}
	}
	// A phase untouched by this mutation (Phase 1) must still be present with
	// its features intact — not otherwise reformatted away.
	containsAll(t, body, `"feature-a"`, `"feature-b"`, `"feature-c"`)
}

// Scenario: a second mutation produces a diff confined to the lines it changed
func TestRsh_SecondMutationDiffIsConfinedToChangedLines(t *testing.T) {
	bin := buildCent(t)
	dir := rshRepo(t, rshBaseRoadmap)
	if out, code := runCent(t, bin, dir, "roadmap", "defer", "seed-thing", "--summary", "x"); code != 0 {
		t.Fatalf("defer exit=%d\n%s", code, out)
	}

	if out, code := runCent(t, bin, dir, "roadmap", "edit", "feature-a", "--description", "Rewritten."); code != 0 {
		t.Fatalf("edit exit=%d\n%s", code, out)
	}
	diff := rshGitOut(t, dir, "diff", "HEAD~1", "HEAD", "--", ".workflow/roadmap.json")
	changedLines := 0
	for _, ln := range strings.Split(diff, "\n") {
		if strings.HasPrefix(ln, "+") || strings.HasPrefix(ln, "-") {
			if strings.HasPrefix(ln, "+++") || strings.HasPrefix(ln, "---") {
				continue
			}
			changedLines++
		}
	}
	if changedLines != 2 { // one removed, one added — the single edited feature
		t.Fatalf("want a 1-line diff (2 +/- lines), got %d:\n%s", changedLines, diff)
	}
}

// Scenario: rendering is idempotent
func TestRsh_RenderingIsIdempotent(t *testing.T) {
	bin := buildCent(t)
	roadmapWithUnknown := `{"custom":{"nested":true},"phases":[` +
		`{"name":"Phase 1","features":[{"name":"feature-a","description":"First thing."}]}]}`
	dir := rshRepo(t, roadmapWithUnknown)

	if out, code := runCent(t, bin, dir, "roadmap", "edit", "feature-a"); code != 0 { // pure no-op
		t.Fatalf("no-op edit exit=%d\n%s", code, out)
	}
	after := mustRead(t, dir+"/.workflow/roadmap.json")
	before := roadmapWithUnknown
	if after != before {
		t.Fatalf("a no-op mutation must leave the file byte-identical:\nbefore: %q\nafter:  %q", before, after)
	}
	if !strings.Contains(after, `"custom":{"nested":true}`) {
		t.Fatalf("unknown top-level fields must survive the round trip:\n%s", after)
	}
}
