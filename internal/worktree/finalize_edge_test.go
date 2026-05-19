package worktree_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

// Worktree removed out-of-band: Remove is idempotent, so APPLY still
// clears the marker and reports finalized.
func TestResolveMerge_WorktreeGoneStillFinalizes(t *testing.T) {
	repo := resolveRepo(t, "sigma")
	writeMarker(t, repo, "sigma")
	if err := worktree.Remove(repo, "sigma", true); err != nil {
		t.Fatalf("pre-remove worktree: %v", err)
	}
	res, err := worktree.ResolveMerge(repo, "sigma", okValidator("complete"))
	if err != nil {
		t.Fatalf("ResolveMerge with worktree gone: %v", err)
	}
	if !res.Finalized {
		t.Fatalf("APPLY must still finalize when worktree already gone: %+v", res)
	}
	if _, err := os.Stat(worktree.PendingPath(repo, "sigma")); !os.IsNotExist(err) {
		t.Fatalf("marker must be cleared: %v", err)
	}
}

// Escalation detail includes the steward report and the proposed diff
// sibling when present (read relative to cwd).
func TestResolveMerge_EscalationDetailIncludesDiff(t *testing.T) {
	repo := resolveRepo(t, "kappa")
	writeMarker(t, repo, "kappa")
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(repo)       //nolint:errcheck
	_ = os.MkdirAll(".workflow", 0o755)
	_ = os.WriteFile(filepath.Join(".workflow", "kappa-merge-steward.md"),
		[]byte("steward report body"), 0o644)
	_ = os.WriteFile(filepath.Join(".workflow", "kappa-merge-steward.diff"),
		[]byte("PROPOSED-DIFF-MARKER"), 0o644)
	res, err := worktree.ResolveMerge(".", "kappa", okValidator("user"))
	if err != nil {
		t.Fatalf("ResolveMerge: %v", err)
	}
	if !strings.Contains(res.EscalationNote, "steward report body") {
		t.Fatalf("escalation note must include report: %q", res.EscalationNote)
	}
	if !strings.Contains(res.EscalationNote, "PROPOSED-DIFF-MARKER") {
		t.Fatalf("escalation note must include proposed diff: %q", res.EscalationNote)
	}
}

func TestStewardDirective_NamesPromptFeatureAndResume(t *testing.T) {
	o := worktree.MergeOutcome{Feature: "delta", TextConflict: true}
	d := o.StewardDirective()
	for _, want := range []string{
		"CENTINELA DIRECTIVE:", worktree.StewardPromptPath, `"delta"`,
		"git-text-conflict", "centinela merge --continue delta",
		"delta-merge-steward.md", "delta-merge-steward.json",
	} {
		if !strings.Contains(d, want) {
			t.Fatalf("directive missing %q: %s", want, d)
		}
	}
}
