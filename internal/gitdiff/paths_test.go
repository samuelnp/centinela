package gitdiff

import (
	"sort"
	"testing"
)

// TestSetPaths covers Set.Paths: it returns the slash-normalized changed paths
// (empty entries dropped) and is nil-receiver safe.
func TestSetPaths(t *testing.T) {
	s := NewSet([]string{"b/y.go", "a/x.go", ""})
	got := s.Paths()
	if len(got) != 2 {
		t.Fatalf("expected 2 paths, got %d: %v", len(got), got)
	}
	sort.Strings(got)
	if got[0] != "a/x.go" || got[1] != "b/y.go" {
		t.Fatalf("unexpected paths: %v", got)
	}

	var nilSet *Set
	if nilSet.Paths() != nil {
		t.Fatal("nil receiver Paths must return nil")
	}
}
