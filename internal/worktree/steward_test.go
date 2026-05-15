package worktree_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

func TestStewardHint_NamesMarkdownAndJSON(t *testing.T) {
	out := worktree.MergeOutcome{Feature: "alpha"}
	if hint := out.StewardHint(); !strings.HasSuffix(hint, "alpha-merge-steward.md") {
		t.Fatalf("StewardHint = %q, want suffix alpha-merge-steward.md", hint)
	}
	if jp := out.StewardJSONPath(); !strings.HasSuffix(jp, "alpha-merge-steward.json") {
		t.Fatalf("StewardJSONPath = %q, want suffix alpha-merge-steward.json", jp)
	}
}

func TestStewardReason_AllCategories(t *testing.T) {
	cases := []struct {
		name string
		out  worktree.MergeOutcome
		want string
	}{
		{"text-conflict", worktree.MergeOutcome{TextConflict: true}, "git-text-conflict"},
		{"validate-fail", worktree.MergeOutcome{ValidateFail: true}, "post-merge-validate-failed"},
		{"clean", worktree.MergeOutcome{}, ""},
	}
	for _, tc := range cases {
		if got := tc.out.StewardReason(); got != tc.want {
			t.Fatalf("%s: StewardReason = %q, want %q", tc.name, got, tc.want)
		}
	}
}
