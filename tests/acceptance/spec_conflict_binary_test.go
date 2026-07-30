package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/parallel-feature-worktrees.feature
// Scenario: Superseding and identical specs never block a merge
// Regression, binary-driven: replays the production failure shape end-to-end
// through the real `centinela merge` command — several worktrees carrying
// byte-identical specs, including two companion scenarios sharing one Given,
// used to block every merge before this hotfix.
func TestAcceptance_SpecConflict_IdenticalSpecsAcrossWorktreesDoNotBlockRealMerge(t *testing.T) {
	work := t.TempDir()
	bin := buildCentinela(t, work)
	repo := seedSpecRepo(t, "dashboard")

	eta := addWorktreeBranch(t, repo, "eta")
	mustWrite(t, filepath.Join(eta, "eta-feature.txt"), "eta work\n")
	commit(t, eta, "eta: unrelated feature work")
	addWorktreeBranch(t, repo, "bystander") // idle checkout, main's specs verbatim

	out, err := runBin(t, bin, repo, "merge", "eta")
	if err != nil {
		t.Fatalf("identical specs across worktrees must not block a real merge: %v\n%s", err, out)
	}
	if strings.Contains(out, "spec conflict") {
		t.Fatalf("merge output must not mention a spec conflict:\n%s", out)
	}
	if _, e := os.Stat(eta); !os.IsNotExist(e) {
		t.Fatalf("eta worktree should be removed after a real clean merge; err=%v", e)
	}
	if _, e := os.Stat(filepath.Join(repo, ".worktrees", "bystander")); e != nil {
		t.Fatalf("bystander worktree must be untouched by eta's merge: %v", e)
	}
}

// Acceptance: specs/parallel-feature-worktrees.feature
// Scenario: Spec conflict across in-flight worktrees is detected before merging
// Regression, binary-driven: two worktrees genuinely diverging on the same
// (file, scenario) must block the real merge command, name both worktrees,
// report the conflict once per print, and never reproduce the observed
// 720KB unbounded report.
func TestAcceptance_SpecConflict_TwoWorktreesDivergeBlocksRealMergeOnce(t *testing.T) {
	work := t.TempDir()
	bin := buildCentinela(t, work)
	repo := seedSpecRepo(t, "checkout") // main's original baseline outcome

	zeta := addWorktreeBranch(t, repo, "zeta")
	mustWrite(t, filepath.Join(zeta, "specs", "login.feature"),
		"Feature: Login\n  Scenario: clash\n    Given user has account\n"+
			"    When user logs in\n    Then app routes to dashboard\n")
	commit(t, zeta, "zeta: resolve clash to dashboard")

	eta := addWorktreeBranch(t, repo, "eta")
	mustWrite(t, filepath.Join(eta, "specs", "login.feature"),
		"Feature: Login\n  Scenario: clash\n    Given user has account\n"+
			"    When user logs in\n    Then app routes to onboarding\n")
	commit(t, eta, "eta: resolve clash to onboarding")

	before := mainHeadSHA(t, repo)
	out, err := runBin(t, bin, repo, "merge", "zeta")
	if err == nil {
		t.Fatalf("diverging worktrees must block the real merge:\n%s", out)
	}
	for _, want := range []string{"spec conflicts", "zeta", "eta", "clash"} {
		if !strings.Contains(out, want) {
			t.Fatalf("blocked output should mention %q:\n%s", want, out)
		}
	}
	// The framework prints the RunE error twice (cobra's own "Error:" line
	// plus main()'s fallback print); each print must carry exactly one
	// conflict entry, never one per duplicate scenario copy or per pairing.
	arrows := strings.Count(out, "\u2194")
	prints := strings.Count(out, "spec conflicts block merge:")
	if prints == 0 || arrows != prints {
		t.Fatalf("expected one conflict entry per error print (arrows=%d prints=%d):\n%s", arrows, prints, out)
	}
	const sanityCap = 4096 // guards the 720KB unbounded-report regression
	if len(out) > sanityCap {
		t.Fatalf("blocked merge output must stay bounded (%d bytes > %d cap):\n%s", len(out), sanityCap, out)
	}
	if _, e := os.Stat(zeta); e != nil {
		t.Fatalf("zeta worktree must be kept when the pre-check blocks: %v", e)
	}
	if got := mainHeadSHA(t, repo); got != before {
		t.Fatalf("main HEAD must not advance when the pre-check blocks: before=%s after=%s", before, got)
	}
}
