package gitutil

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// THE test this whole feature exists for: a mutation commit must not swallow
// another session's staged work, and must leave it staged.
func TestCommitPathsLeavesUnrelatedStagedWorkStaged(t *testing.T) {
	gitAvailable(t)
	dir := newRepo(t)
	write(t, dir, "internal/example/thing.go", "package example\n")
	mustGit(t, dir, "add", "internal/example/thing.go")
	write(t, dir, "README.md", "dirty\n")
	write(t, dir, ".workflow/roadmap.json", "{}\n")
	write(t, dir, "ROADMAP.md", "# Roadmap\n")

	if err := CommitPaths(dir, "chore(roadmap): defer x", []string{".workflow/roadmap.json", "ROADMAP.md"}); err != nil {
		t.Fatalf("CommitPaths: %v", err)
	}
	got := strings.Join(commitPaths(t, dir, "HEAD"), ",")
	if got != ".workflow/roadmap.json,ROADMAP.md" {
		t.Fatalf("commit changed %q, want only the roadmap-state pathspec", got)
	}
	staged := mustGit(t, dir, "diff", "--cached", "--name-only")
	if strings.TrimSpace(staged) != "internal/example/thing.go" {
		t.Fatalf("staged after commit = %q, want the unrelated file still staged", staged)
	}
	if !strings.Contains(mustGit(t, dir, "status", "--porcelain"), "README.md") {
		t.Fatal("the unstaged dirty file must survive untouched")
	}
}

func TestCommitPathsExactlyOneCommitWithTheGivenMessage(t *testing.T) {
	gitAvailable(t)
	dir := newRepo(t)
	before := headCount(t, dir)
	write(t, dir, ".workflow/roadmap.json", "{}\n")
	if err := CommitPaths(dir, "chore(roadmap): defer x", []string{".workflow/roadmap.json", "ROADMAP.md"}); err != nil {
		t.Fatalf("CommitPaths: %v", err)
	}
	if headCount(t, dir) == before {
		t.Fatal("no commit was created")
	}
	if got := strings.TrimSpace(mustGit(t, dir, "log", "-1", "--pretty=%s")); got != "chore(roadmap): defer x" {
		t.Fatalf("message = %q", got)
	}
	// The absent ROADMAP.md was dropped from the pathspec, not fatal.
	if got := strings.Join(commitPaths(t, dir, "HEAD"), ","); got != ".workflow/roadmap.json" {
		t.Fatalf("commit changed %q", got)
	}
}

func TestCommitPathsNoChangeMakesNoCommit(t *testing.T) {
	gitAvailable(t)
	dir := newRepo(t)
	write(t, dir, ".workflow/roadmap.json", "{}\n")
	if err := CommitPaths(dir, "chore(roadmap): defer x", []string{".workflow/roadmap.json"}); err != nil {
		t.Fatal(err)
	}
	before := headCount(t, dir)
	err := CommitPaths(dir, "chore(roadmap): defer x again", []string{".workflow/roadmap.json"})
	if !errors.Is(err, ErrNothingToCommit) {
		t.Fatalf("want ErrNothingToCommit, got %v", err)
	}
	if headCount(t, dir) != before {
		t.Fatal("an unchanged pathspec must create no empty commit")
	}
}

func TestCommitPathsAllPathsAbsentIsNothingToCommit(t *testing.T) {
	gitAvailable(t)
	dir := newRepo(t)
	if err := CommitPaths(dir, "m", []string{"nope.json"}); !errors.Is(err, ErrNothingToCommit) {
		t.Fatalf("want ErrNothingToCommit, got %v", err)
	}
}

func TestCommitPathsOnANonRepoWritesNothing(t *testing.T) {
	gitAvailable(t)
	dir := t.TempDir()
	write(t, dir, ".workflow/roadmap.json", "{}\n")
	err := CommitPaths(dir, "m", []string{".workflow/roadmap.json"})
	if err == nil || !strings.Contains(err.Error(), "no git repository") {
		t.Fatalf("want a no-repository error, got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, ".git")); statErr == nil {
		t.Fatal("CommitPaths must never create a repository")
	}
}
