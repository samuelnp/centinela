package treestate

import "testing"

const srcDiff = "diff --git a/internal/x.go b/internal/x.go\n@@ -1 +1 @@\n-a\n+b\n"

const wfDiff = "diff --git a/.workflow/f-gatekeeper.md b/.workflow/f-gatekeeper.md\n@@ -1 +1 @@\n-x\n+y\n"

// The single most important assertion in this feature (D3a): the verifier
// writing its own report must NOT invalidate the stamp it just took, while a
// tracked source edit must.
func TestDigestIgnoresWorkflowOnlyChurn(t *testing.T) {
	clean := Digest("", "", nil)
	churn := Digest(" M .workflow/f-gatekeeper.md\n?? .workflow/f-gatekeeper.json\n", wfDiff, nil)
	if clean != churn {
		t.Fatalf(".workflow/-only churn staled the digest: %q vs %q", clean, churn)
	}
}

func TestDigestChangesOnTrackedSource(t *testing.T) {
	clean := Digest("", "", nil)
	dirty := Digest(" M internal/x.go\n", srcDiff, nil)
	if clean == dirty {
		t.Fatal("tracked source edit must stale the digest")
	}
}

func TestDigestKeepsSourceWhenWorkflowAlsoChanged(t *testing.T) {
	both := Digest(" M internal/x.go\n M .workflow/f.json\n", srcDiff+wfDiff, nil)
	only := Digest(" M internal/x.go\n", srcDiff, nil)
	if both != only {
		t.Fatalf("mixed churn must reduce to the source change: %q vs %q", both, only)
	}
}

func TestDigestIsOrderIndependent(t *testing.T) {
	a := Digest(" M b.go\n M a.go\n", "", nil)
	b := Digest(" M a.go\n M b.go\n", "", nil)
	if a != b {
		t.Fatal("status ordering must not change the digest")
	}
}

func TestDigestKeepsRenameOutOfWorkflow(t *testing.T) {
	moved := Digest("R  .workflow/x.md -> docs/x.md\n", "", nil)
	if moved == Digest("", "", nil) {
		t.Fatal("a rename out of .workflow/ is a real change")
	}
}

func TestDigestIsPrefixed(t *testing.T) {
	if got := Digest("", "", nil); got[:7] != "sha256:" {
		t.Fatalf("want sha256: prefix, got %q", got)
	}
}
