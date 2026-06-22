package reconstruct

import (
	"regexp"
	"strings"
	"testing"
)

// realScenarioLine mirrors internal/gates.spec_traceability_parse.scenarioLine
// exactly so this test proves generated features satisfy the real parser shape
// (a Feature: line + ≥1 indented Scenario: line) without importing the
// unexported gates parser.
var realScenarioLine = regexp.MustCompile(`^\s+Scenario(?: Outline)?:\s*(.+?)\s*$`)

func parsesWithRealParser(body string) bool {
	hasFeature, hasScenario := false, false
	for _, ln := range strings.Split(body, "\n") {
		if strings.HasPrefix(ln, "Feature:") {
			hasFeature = true
		}
		if realScenarioLine.MatchString(ln) {
			hasScenario = true
		}
	}
	return hasFeature && hasScenario
}

func TestFeatureSkeleton_ParsesAndCountsTodos(t *testing.T) {
	for _, role := range []Role{RoleCommand, RoleEndpoint, RoleModule, "", "weird"} {
		body, todos := featureSkeleton(Target{Pkg: "internal/x", Slug: "internal-x", Role: role, Reason: "r"})
		if !parsesWithRealParser(body) {
			t.Fatalf("role %q: generated feature must parse with real parser:\n%s", role, body)
		}
		if todos != 3 || strings.Count(body, todoMarker) != 3 {
			t.Fatalf("role %q: expected 3 TODO markers, got %d", role, todos)
		}
	}
}

func TestFeatureSkeleton_NoFabricatedSteps(t *testing.T) {
	body, _ := featureSkeleton(Target{Pkg: "p", Slug: "p", Role: RoleModule, Reason: "r"})
	for _, ln := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(ln)
		isStep := strings.HasPrefix(trimmed, "Given ") ||
			strings.HasPrefix(trimmed, "When ") || strings.HasPrefix(trimmed, "Then ")
		if isStep && !strings.Contains(ln, todoMarker) {
			t.Fatalf("step line asserts fabricated behavior (no TODO): %q", ln)
		}
	}
}

func TestFeatureSkeleton_GoldenFragment(t *testing.T) {
	body, _ := featureSkeleton(Target{Pkg: "cmd/app", Slug: "cmd-app", Role: RoleCommand, Reason: "command surface"})
	want := "Feature: cmd-app — reconstructed command behavior\n"
	if !strings.HasPrefix(body, want) {
		t.Fatalf("golden header mismatch:\n%s", body)
	}
	if !strings.Contains(body, "  Scenario: the command performs its primary behavior\n") {
		t.Fatalf("golden scenario line missing:\n%s", body)
	}
}

func TestNarrativeFor(t *testing.T) {
	if !strings.Contains(narrativeFor(RoleEndpoint), "request/response") ||
		!strings.Contains(narrativeFor(RoleModule), "module") ||
		!strings.Contains(narrativeFor(RoleCommand), "command") {
		t.Fatal("narrativeFor wrong per role")
	}
}

func TestRoleOrModule(t *testing.T) {
	if roleOrModule("") != RoleModule || roleOrModule(RoleCommand) != RoleCommand {
		t.Fatal("roleOrModule normalization wrong")
	}
}
