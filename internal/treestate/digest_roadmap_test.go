package treestate

import "testing"

const mdDiff = "diff --git a/ROADMAP.md b/ROADMAP.md\n@@ -1 +1 @@\n-x\n+y\n"

// The disable_auto_commit case: a regenerated but uncommitted ROADMAP.md sits
// at the repo root, outside .workflow/, and must not stale the stamp.
func TestDigestIgnoresRegeneratedRoadmapMarkdown(t *testing.T) {
	clean := Digest("", "", nil)
	churn := Digest(" M ROADMAP.md\n M .workflow/roadmap.json\n", mdDiff, nil)
	if clean != churn {
		t.Fatalf("roadmap-state-only churn staled the digest: %q vs %q", clean, churn)
	}
}

// The exemption is a strict subset test, never a filter: a source edit hiding
// alongside roadmap churn must still stale the stamp (R1).
func TestDigestKeepsSourceWhenRoadmapStateAlsoChanged(t *testing.T) {
	mixed := Digest(" M internal/x.go\n M ROADMAP.md\n", srcDiff+mdDiff, nil)
	if mixed == Digest("", "", nil) {
		t.Fatal("a source edit must stale the digest even beside roadmap churn")
	}
	if mixed != Digest(" M internal/x.go\n", srcDiff, nil) {
		t.Fatal("mixed churn must reduce to exactly the source change")
	}
}

// A lookalike path outside the roadmap-state set is ordinary product.
func TestDigestDoesNotExemptRoadmapLookalikes(t *testing.T) {
	for _, status := range []string{" M docs/ROADMAP.md\n", " M ROADMAP.md.bak\n", " M workflow/roadmap.json\n"} {
		if Digest(status, "", nil) == Digest("", "", nil) {
			t.Fatalf("%q must not be exempt", status)
		}
	}
}

func TestUntrackedPathsSkipsRoadmapState(t *testing.T) {
	got := UntrackedPaths("?? ROADMAP.md\n?? .workflow/roadmap.json\n?? internal/new.go\n")
	if len(got) != 1 || got[0] != "internal/new.go" {
		t.Fatalf("UntrackedPaths = %v, want only the source file", got)
	}
}
