package gitdiff

import (
	"errors"
	"strings"
	"testing"
)

func TestChangedFiles_DegradesOnMissingBase(t *testing.T) {
	s := &stub{responses: map[string]stubReply{
		"git merge-base HEAD main": {err: errors.New("fatal: Not a valid object name main")},
	}}
	set, sum, err := newResolver(s).ChangedFiles("main", true)
	if err != nil || set != nil {
		t.Fatalf("expected nil set and nil error on degrade")
	}
	if !strings.Contains(sum.Degrade, "diff base \"main\" not found") {
		t.Fatalf("expected base-not-found degrade reason, got %q", sum.Degrade)
	}
}

func TestChangedFiles_DegradesOnNonGitRepo(t *testing.T) {
	s := &stub{responses: map[string]stubReply{
		"git merge-base HEAD main": {err: errors.New("fatal: not a git repository (or any of the parent directories): .git")},
	}}
	_, sum, _ := newResolver(s).ChangedFiles("main", true)
	if sum.Degrade != "not a git repository" {
		t.Fatalf("expected non-git-repo degrade, got %q", sum.Degrade)
	}
}

func TestChangedFiles_DegradesOnDiffFailure(t *testing.T) {
	s := &stub{responses: map[string]stubReply{
		"git merge-base HEAD main":                    {out: "abc\n"},
		"git diff --name-only --diff-filter=ACMR abc": {err: errors.New("boom")},
	}}
	_, sum, _ := newResolver(s).ChangedFiles("main", true)
	if !strings.Contains(sum.Degrade, "git diff failed") {
		t.Fatalf("expected diff-failed degrade, got %q", sum.Degrade)
	}
}

func TestChangedFiles_DegradesOnLsFilesFailure(t *testing.T) {
	s := &stub{responses: map[string]stubReply{
		"git merge-base HEAD main":                    {out: "abc\n"},
		"git diff --name-only --diff-filter=ACMR abc": {out: ""},
		"git ls-files --others --exclude-standard":    {err: errors.New("boom")},
	}}
	_, sum, _ := newResolver(s).ChangedFiles("main", true)
	if !strings.Contains(sum.Degrade, "git ls-files failed") {
		t.Fatalf("expected ls-files-failed degrade, got %q", sum.Degrade)
	}
}

func TestChangedFiles_GenericMergeBaseFailureSurfacesMessage(t *testing.T) {
	s := &stub{responses: map[string]stubReply{
		"git merge-base HEAD main": {err: errors.New("unexpected: stderr noise")},
	}}
	_, sum, _ := newResolver(s).ChangedFiles("main", true)
	if !strings.Contains(sum.Degrade, "git merge-base failed") {
		t.Fatalf("expected generic merge-base failure, got %q", sum.Degrade)
	}
}
