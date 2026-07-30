package orchestration

import (
	"os"
	"strings"
	"testing"
)

// seedBriefs chdirs into a temp repo containing n sibling briefs, so a
// reintroduced filesystem glob would visibly grow the required set.
func seedBriefs(t *testing.T, n int) {
	t.Helper()
	t.Chdir(t.TempDir())
	if err := os.MkdirAll("docs/features", 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < n; i++ {
		p := "docs/features/other-" + string(rune('a'+i%26)) + string(rune('a'+i/26)) + ".md"
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRequiredPlanInputsReturnsExactlyOwnBriefAndPlan(t *testing.T) {
	seedBriefs(t, 5)
	got := RequiredPlanInputs("token-diet")
	want := []string{"docs/features/token-diet.md", "docs/plans/token-diet.md"}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("got %v want %v", got, want)
	}
}

// The no-glob invariant: identical results with 0 briefs and with 50.
func TestRequiredPlanInputsDoesNotGrowWithBriefCount(t *testing.T) {
	seedBriefs(t, 0)
	empty := RequiredPlanInputs("token-diet")
	seedBriefs(t, 50)
	many := RequiredPlanInputs("token-diet")
	if strings.Join(empty, "|") != strings.Join(many, "|") {
		t.Fatalf("set grew with brief count: %v vs %v", empty, many)
	}
}

func TestRequiredPlanInputsIsConstructionNotExistence(t *testing.T) {
	seedBriefs(t, 0)
	got := RequiredPlanInputs("ghost")
	if len(got) != 2 {
		t.Fatalf("ghost feature must still require both paths, got %v", got)
	}
}

func TestValidatePlanSnapshotInputsAcceptsSuperset(t *testing.T) {
	seedBriefs(t, 0)
	inputs := []string{"docs/features/token-diet.md", "docs/plans/token-diet.md"}
	for i := 0; i < 120; i++ {
		inputs = append(inputs, "docs/features/legacy-"+string(rune('a'+i%26))+".md")
	}
	if err := validatePlanSnapshotInputs("p", "token-diet", "plan", RolePlanner, inputs); err != nil {
		t.Fatalf("superset must validate, got %v", err)
	}
}

func TestValidatePlanSnapshotInputsNamesEachMissingPath(t *testing.T) {
	seedBriefs(t, 0)
	err := validatePlanSnapshotInputs("p", "td", "plan", RolePlanner, []string{"docs/features/td.md"})
	if err == nil || !strings.Contains(err.Error(), "docs/plans/td.md") {
		t.Fatalf("expected missing plan path, got %v", err)
	}
	err = validatePlanSnapshotInputs("p", "td", "plan", RoleBigThinker, nil)
	if err == nil || !strings.Contains(err.Error(), "missing feature-doc snapshot inputs") ||
		!strings.Contains(err.Error(), "docs/features/td.md") || !strings.Contains(err.Error(), "docs/plans/td.md") {
		t.Fatalf("empty inputs must name both paths, got %v", err)
	}
	if err := validatePlanSnapshotInputs("p", "td", "code", RoleSeniorEngineer, nil); err != nil {
		t.Fatalf("non-plan step must bypass, got %v", err)
	}
}

func TestNormalizeFeatureDocPathIsSymmetric(t *testing.T) {
	cases := map[string]string{
		"docs/features/td.md":                 "docs/features/td.md",
		"./docs/features/td.md":               "docs/features/td.md",
		`docs\features\td.md`:                 "docs/features/td.md",
		"/home/user/repo/docs/features/td.md": "docs/features/td.md",
		"docs/plans/td.md":                    "docs/plans/td.md",
		"./docs/plans/td.md":                  "docs/plans/td.md",
		`docs\plans\td.md`:                    "docs/plans/td.md",
		"/home/user/repo/docs/plans/td.md":    "docs/plans/td.md",
		"  docs/plans/td.md  ":                "docs/plans/td.md",
		"notes/other.md":                      "notes/other.md",
	}
	for in, want := range cases {
		if got := normalizeFeatureDocPath(in); got != want {
			t.Errorf("normalize(%q) = %q, want %q", in, got, want)
		}
	}
}
