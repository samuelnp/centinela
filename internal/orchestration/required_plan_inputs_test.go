package orchestration

import (
	"os"
	"sort"
	"testing"
)

// chdirRPI moves into a tempdir seeded with the given feature briefs so a
// reintroduced docs/features glob would show up as extra required entries.
func chdirRPI(t *testing.T, briefs ...string) {
	t.Helper()
	t.Chdir(t.TempDir())
	if err := os.MkdirAll("docs/features", 0o755); err != nil {
		t.Fatal(err)
	}
	for _, b := range briefs {
		if err := os.WriteFile("docs/features/"+b, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRequiredPlanInputsIsOwnBriefAndPlanOnly(t *testing.T) {
	chdirRPI(t, "demo.md", "alpha.md", "beta.md")
	got := RequiredPlanInputs("demo")
	for _, want := range []string{"docs/features/demo.md", "docs/plans/demo.md"} {
		if !contains(got, want) {
			t.Fatalf("RequiredPlanInputs missing %q in %v", want, got)
		}
	}
	for _, sibling := range []string{"docs/features/alpha.md", "docs/features/beta.md"} {
		if contains(got, sibling) {
			t.Fatalf("sibling brief %q leaked into the required set: %v", sibling, got)
		}
	}
	if !sort.StringsAreSorted(got) {
		t.Fatalf("RequiredPlanInputs not sorted: %v", got)
	}
}

func TestRequiredPlanInputsDoesNotDuplicateBriefOnDisk(t *testing.T) {
	// demo.md exists on disk AND is derived by construction — must appear once.
	chdirRPI(t, "demo.md")
	got := RequiredPlanInputs("demo")
	n := 0
	for _, g := range got {
		if g == "docs/features/demo.md" {
			n++
		}
	}
	if n != 1 || len(got) != 2 {
		t.Fatalf("expected demo brief once in a 2-entry set, got %d in %v", n, got)
	}
}

func TestRequiredPlanInputsNormalizesPaths(t *testing.T) {
	chdirRPI(t, "demo.md")
	for _, p := range RequiredPlanInputs("demo") {
		if p != normalizeFeatureDocPath(p) {
			t.Fatalf("entry not normalized: %q", p)
		}
	}
	if normalizeFeatureDocPath(`.\docs\features\demo.md`) != "docs/features/demo.md" {
		t.Fatalf("normalizeFeatureDocPath did not strip backslash/dot prefix")
	}
	if normalizeFeatureDocPath(`.\docs\plans\demo.md`) != "docs/plans/demo.md" {
		t.Fatalf("normalizeFeatureDocPath did not anchor on docs/plans/")
	}
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
