package gitdiff

import (
	"errors"
	"strings"
	"testing"
)

// notAnObject is the error real git returns for an unresolvable ref name.
var notAnObject = errors.New("exit status 128: fatal: Not a valid object name main")

// recordingResolver answers merge-base for exactly the refs in resolves and
// records every ref it was asked about, in order.
func recordingResolver(resolves map[string]string, tried *[]string) *Resolver {
	return &Resolver{Run: func(_ string, args ...string) (string, error) {
		switch args[0] {
		case "merge-base":
			ref := args[2]
			*tried = append(*tried, ref)
			if sha, ok := resolves[ref]; ok {
				return sha + "\n", nil
			}
			return "", notAnObject
		case "diff":
			return "internal/a.go\n", nil
		}
		return "", nil
	}}
}

// Finding 1 regression: actions/checkout leaves a detached HEAD with no local
// branch ref for the default branch, so `git merge-base HEAD main` exits 128
// even on a full-history clone. Without the origin/ retry every diff-aware
// gate silently degrades in CI, and a gate that never runs looks exactly like
// a gate that always passes.
func TestChangedFiles_FallsBackToTheRemoteTrackingRef(t *testing.T) {
	var tried []string
	r := recordingResolver(map[string]string{"origin/main": "abc123"}, &tried)
	set, summary, err := r.ChangedFiles("main", false)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if summary.Degrade != "" {
		t.Fatalf("must not degrade when origin/main resolves: %q", summary.Degrade)
	}
	if set == nil || !set.Contains("internal/a.go") {
		t.Fatalf("set = %v", set)
	}
	if strings.Join(tried, ",") != "main,origin/main" {
		t.Fatalf("tried = %v; the bare name must be preferred, then origin/", tried)
	}
	if summary.Base != "origin/main" {
		t.Fatalf("Base = %q; the header must name the ref that actually resolved", summary.Base)
	}
}

func TestChangedFiles_PrefersTheLocalRefAndDoesNotRetry(t *testing.T) {
	var tried []string
	r := recordingResolver(map[string]string{"main": "abc123"}, &tried)
	_, summary, _ := r.ChangedFiles("main", false)
	if len(tried) != 1 || tried[0] != "main" {
		t.Fatalf("a resolving local ref must not trigger a retry: %v", tried)
	}
	if summary.Base != "main" {
		t.Fatalf("Base = %q", summary.Base)
	}
}

func TestChangedFiles_QualifiedBaseIsNeverRewritten(t *testing.T) {
	var tried []string
	r := recordingResolver(map[string]string{}, &tried)
	_, summary, _ := r.ChangedFiles("upstream/release", false)
	if len(tried) != 1 {
		t.Fatalf("an already-qualified base must not be prefixed again: %v", tried)
	}
	if !strings.Contains(summary.Degrade, `"upstream/release" not found`) ||
		strings.Contains(summary.Degrade, "also tried") {
		t.Fatalf("degrade = %q", summary.Degrade)
	}
}

func TestChangedFiles_DegradeNamesEveryRefFormTried(t *testing.T) {
	var tried []string
	r := recordingResolver(map[string]string{}, &tried)
	_, summary, _ := r.ChangedFiles("main", false)
	if strings.Join(tried, ",") != "main,origin/main" {
		t.Fatalf("tried = %v", tried)
	}
	if !strings.Contains(summary.Degrade, `diff base "main" not found (also tried "origin/main")`) {
		t.Fatalf("degrade must not imply only one form was attempted: %q", summary.Degrade)
	}
}
