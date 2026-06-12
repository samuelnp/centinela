package orchestration

import (
	"os"
	"sort"
	"testing"
)

// chdirRPI moves into a tempdir seeded with the given feature briefs (plus a
// docs/plans dir) so RequiredPlanInputs globs a controlled docs/features set.
func chdirRPI(t *testing.T, briefs ...string) {
	t.Helper()
	d := t.TempDir()
	o, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(o) })
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("docs/features", 0o755); err != nil {
		t.Fatal(err)
	}
	for _, b := range briefs {
		if err := os.WriteFile("docs/features/"+b, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRequiredPlanInputsIncludesBriefPlanAndAllFeatures(t *testing.T) {
	chdirRPI(t, "demo.md", "alpha.md", "beta.md")
	got := RequiredPlanInputs("demo")
	for _, want := range []string{
		"docs/features/demo.md",
		"docs/features/alpha.md",
		"docs/features/beta.md",
		"docs/plans/demo.md",
	} {
		if !contains(got, want) {
			t.Fatalf("RequiredPlanInputs missing %q in %v", want, got)
		}
	}
	if !sort.StringsAreSorted(got) {
		t.Fatalf("RequiredPlanInputs not sorted: %v", got)
	}
}

func TestRequiredPlanInputsDedupsBriefAlreadyOnDisk(t *testing.T) {
	// demo.md exists on disk AND is added explicitly — must appear once.
	chdirRPI(t, "demo.md")
	got := RequiredPlanInputs("demo")
	n := 0
	for _, g := range got {
		if g == "docs/features/demo.md" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("expected demo brief once, got %d in %v", n, got)
	}
}

func TestRequiredPlanInputsNormalizesPaths(t *testing.T) {
	// Glob yields clean slash paths; assert the plan path is normalized and the
	// set has no "./" or backslash residue.
	chdirRPI(t, "demo.md")
	for _, p := range RequiredPlanInputs("demo") {
		if p != normalizeFeatureDocPath(p) {
			t.Fatalf("entry not normalized: %q", p)
		}
	}
	if normalizeFeatureDocPath(`.\docs\features\demo.md`) != "docs/features/demo.md" {
		t.Fatalf("normalizeFeatureDocPath did not strip backslash/dot prefix")
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
