package gitdiff

import (
	"errors"
	"testing"
)

func TestChangedFilesStaged_HappyPath(t *testing.T) {
	s := &stub{responses: map[string]stubReply{
		"git diff --cached --name-only --diff-filter=ACMR": {out: "a.go\nb.go\n"},
	}}
	set, sum, err := newResolver(s).ChangedFilesStaged()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sum.Degrade != "" {
		t.Fatalf("did not expect degrade, got %q", sum.Degrade)
	}
	if sum.Base != "STAGED" || sum.Files != 2 {
		t.Fatalf("unexpected summary: %+v", sum)
	}
	for _, p := range []string{"a.go", "b.go"} {
		if !set.Contains(p) {
			t.Fatalf("expected staged set to contain %q", p)
		}
	}
}

func TestChangedFilesStaged_EmptyIndex(t *testing.T) {
	s := &stub{responses: map[string]stubReply{
		"git diff --cached --name-only --diff-filter=ACMR": {out: ""},
	}}
	set, sum, err := newResolver(s).ChangedFilesStaged()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sum.Degrade != "" || sum.Files != 0 || set.Len() != 0 {
		t.Fatalf("empty index must yield clean empty set: %+v len=%d", sum, set.Len())
	}
}

func TestChangedFilesStaged_DegradesOnNonGitRepo(t *testing.T) {
	s := &stub{responses: map[string]stubReply{
		"git diff --cached --name-only --diff-filter=ACMR": {
			err: errors.New("fatal: not a git repository (or any of the parent directories): .git"),
		},
	}}
	set, sum, err := newResolver(s).ChangedFilesStaged()
	if err != nil {
		t.Fatalf("staged degrade must return nil error, got %v", err)
	}
	if set != nil {
		t.Fatalf("staged degrade must return nil set")
	}
	if sum.Degrade != "not a git repository" {
		t.Fatalf("expected non-git-repo degrade, got %q", sum.Degrade)
	}
}

func TestChangedFilesStaged_DegradesOnGenericFailure(t *testing.T) {
	s := &stub{responses: map[string]stubReply{
		"git diff --cached --name-only --diff-filter=ACMR": {err: errors.New("boom")},
	}}
	_, sum, err := newResolver(s).ChangedFilesStaged()
	if err != nil {
		t.Fatalf("generic failure must still degrade with nil error, got %v", err)
	}
	if sum.Degrade == "" {
		t.Fatalf("expected a non-empty degrade reason for a git failure")
	}
}
