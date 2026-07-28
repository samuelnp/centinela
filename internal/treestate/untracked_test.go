package treestate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUntrackedPathsSkipsWorkflowAndTracked(t *testing.T) {
	status := "?? newpkg/a.go\n M internal/x.go\n?? .workflow/f.json\n?? \"quoted path.go\"\n?? \n"
	got := UntrackedPaths(status)
	want := []string{"newpkg/a.go", "quoted path.go"}
	if len(got) != len(want) {
		t.Fatalf("UntrackedPaths = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("UntrackedPaths = %v, want %v", got, want)
		}
	}
}

func TestHashUntrackedTracksContentAndOrder(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte("one"), 0o644); err != nil {
		t.Fatal(err)
	}
	first := HashUntracked(root, []string{"a.go"})
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte("two"), 0o644); err != nil {
		t.Fatal(err)
	}
	if second := HashUntracked(root, []string{"a.go"}); first[0] == second[0] {
		t.Fatal("an edit to an untracked file must change its hash entry")
	}
	os.WriteFile(filepath.Join(root, "b.go"), []byte("b"), 0o644) //nolint:errcheck
	ab := HashUntracked(root, []string{"a.go", "b.go"})
	ba := HashUntracked(root, []string{"b.go", "a.go"})
	if ab[0] != ba[0] || ab[1] != ba[1] {
		t.Fatal("entries must be sorted, so listing order cannot move the digest")
	}
}

// An unreadable path is a change to the verified tree, never a silent skip.
func TestHashUntrackedMarksUnreadable(t *testing.T) {
	got := HashUntracked(t.TempDir(), []string{"missing.go"})
	if len(got) != 1 || got[0] == "missing.go\x00" {
		t.Fatalf("unreadable path must carry a marker: %q", got)
	}
}

func TestDigestMovesWithUntrackedContent(t *testing.T) {
	status := "?? newpkg/a.go\n"
	before := Digest(status, "", []string{"newpkg/a.go\x00aaa"})
	after := Digest(status, "", []string{"newpkg/a.go\x00bbb"})
	if before == after {
		t.Fatal("untracked content must be part of the digest")
	}
}
