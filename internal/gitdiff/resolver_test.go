package gitdiff

import (
	"errors"
	"strings"
	"testing"
)

type stub struct {
	responses map[string]stubReply
}

type stubReply struct {
	out string
	err error
}

func (s *stub) run(name string, args ...string) (string, error) {
	key := name + " " + strings.Join(args, " ")
	r, ok := s.responses[key]
	if !ok {
		return "", errors.New("unexpected call: " + key)
	}
	return r.out, r.err
}

func newResolver(s *stub) *Resolver {
	return &Resolver{Run: s.run}
}

func TestChangedFiles_HappyPath_WithUntracked(t *testing.T) {
	s := &stub{responses: map[string]stubReply{
		"git merge-base HEAD main":                          {out: "abc123\n"},
		"git diff --name-only --diff-filter=ACMR abc123":    {out: "a.go\nb.go\n"},
		"git ls-files --others --exclude-standard":          {out: "new.go\n"},
	}}
	set, sum, err := newResolver(s).ChangedFiles("", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sum.Degrade != "" {
		t.Fatalf("did not expect degrade, got %q", sum.Degrade)
	}
	if sum.Base != "main" || sum.Files != 3 {
		t.Fatalf("unexpected summary: %+v", sum)
	}
	for _, p := range []string{"a.go", "b.go", "new.go"} {
		if !set.Contains(p) {
			t.Fatalf("expected set to contain %q", p)
		}
	}
}

func TestChangedFiles_SkipsUntrackedWhenDisabled(t *testing.T) {
	s := &stub{responses: map[string]stubReply{
		"git merge-base HEAD master":                       {out: "deadbeef\n"},
		"git diff --name-only --diff-filter=ACMR deadbeef": {out: "x.go\n"},
	}}
	set, sum, err := newResolver(s).ChangedFiles("master", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sum.Base != "master" || sum.Files != 1 || !set.Contains("x.go") {
		t.Fatalf("unexpected result: %+v", sum)
	}
}

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
		"git merge-base HEAD main":                       {out: "abc\n"},
		"git diff --name-only --diff-filter=ACMR abc":    {err: errors.New("boom")},
	}}
	_, sum, _ := newResolver(s).ChangedFiles("main", true)
	if !strings.Contains(sum.Degrade, "git diff failed") {
		t.Fatalf("expected diff-failed degrade, got %q", sum.Degrade)
	}
}

func TestChangedFiles_DegradesOnLsFilesFailure(t *testing.T) {
	s := &stub{responses: map[string]stubReply{
		"git merge-base HEAD main":                       {out: "abc\n"},
		"git diff --name-only --diff-filter=ACMR abc":    {out: ""},
		"git ls-files --others --exclude-standard":       {err: errors.New("boom")},
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

func TestRunGit_WrapsExecErrors(t *testing.T) {
	_, err := runGit("this-binary-does-not-exist-abc123")
	if err == nil {
		t.Fatalf("expected error from missing binary")
	}
}
