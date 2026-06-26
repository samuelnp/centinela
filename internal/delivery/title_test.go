package delivery

import "testing"

// TestComposePRTitleFromChangelogSeed: a conventional-commit stub line becomes
// the PR title verbatim (the changelog summary doubles as the title).
func TestComposePRTitleFromChangelogSeed(t *testing.T) {
	got := ComposePRTitle(Evidence{Feature: "alpha", ChangelogStub: "- feat: add the widget\n"})
	if got != "feat: add the widget" {
		t.Fatalf("title from seed = %q", got)
	}
}

// TestComposePRTitleFallsBackToSlug: with no stub and no brief, the title must
// still be non-empty (gh requires it) — it falls back to the feature slug.
func TestComposePRTitleFallsBackToSlug(t *testing.T) {
	if got := ComposePRTitle(Evidence{Feature: "alpha"}); got != "alpha" {
		t.Fatalf("fallback title = %q, want feature slug", got)
	}
}
