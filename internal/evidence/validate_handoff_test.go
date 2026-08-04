package evidence

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// seedQAEvidence writes structurally complete qa-senior evidence carrying
// handoffTo, so ValidateFeature's verdict turns on the CHAIN and nothing else.
func seedQAEvidence(t *testing.T, feature, handoffTo string) {
	t.Helper()
	if err := os.MkdirAll("tests/unit", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("tests/unit/x_test.go", []byte("package unit\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	edge := ".workflow/" + feature + "-edge-cases.md"
	if err := os.WriteFile(edge, []byte("# edges\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := Skeleton(feature, orchestration.RoleQASeniorEngineer, "0")
	r.Inputs = []string{"docs/features/" + feature + ".md"}
	r.Outputs = []string{"tests/unit/x_test.go", edge}
	r.EdgeCases = []string{"e"}
	r.HandoffTo = handoffTo
	if err := WriteAtomic(feature, orchestration.RoleQASeniorEngineer, r); err != nil {
		t.Fatal(err)
	}
}

// The pre-flight command role prompts tell agents to run must not hand out a
// green light `centinela complete` then contradicts.
func TestValidateFeatureAgreesWithTheCompletionChainGate(t *testing.T) {
	prefillRepo(t, "f", false)
	seedQAEvidence(t, "f", "banana")
	hints := ValidateFeature("f", nil)
	if len(hints) == 0 {
		t.Fatal("an out-of-chain handoffTo must not report evidence ok")
	}
	found := false
	for _, h := range hints {
		if h.Field == "handoffTo" && strings.Contains(h.Message, "gatekeeper") {
			found = true
		}
	}
	if !found {
		t.Fatalf("a handoffTo hint naming the derived successor is required, got %+v", hints)
	}
}

// ...and it must not be STRICTER than the gate either: the derived successor,
// and the successor step's other contract occupant, both stay silent.
func TestValidateFeatureStaysSilentOnInChainHandoffs(t *testing.T) {
	for _, got := range []string{"gatekeeper", "validation-specialist"} {
		prefillRepo(t, "f", false)
		seedQAEvidence(t, "f", got)
		if hints := ValidateFeature("f", nil); len(hints) != 0 {
			t.Fatalf("handoffTo %q is in-chain, got %+v", got, hints)
		}
	}
}

// A bare "<feature>-.json" names no role, so it must be skipped rather than
// parsed as one — otherwise the walk would attribute hints to an empty role.
func TestValidateFeatureSkipsRolelessFilenames(t *testing.T) {
	prefillRepo(t, "f", false)
	if err := os.WriteFile(".workflow/f-.json", []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if hints := ValidateFeature("f", nil); len(hints) != 0 {
		t.Fatalf("a roleless filename must be ignored, got %+v", hints)
	}
}

// The out-of-band merge-steward is never chain-checked here: its rule is the
// literal verdict pair, and previewing a rule the gate does not apply would
// make the pre-flight refuse what completion accepts.
func TestValidateFeatureSkipsRolesTheStepDoesNotRequire(t *testing.T) {
	prefillRepo(t, "f", false)
	r := Skeleton("f", orchestration.RoleMergeSteward, "0")
	r.Inputs = []string{"docs/features/f.md"}
	r.Outputs = []string{".workflow/f-merge-steward.md"}
	r.HandoffTo = "user"
	if err := os.WriteFile(".workflow/f-merge-steward.md", []byte("# steward\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteAtomic("f", orchestration.RoleMergeSteward, r); err != nil {
		t.Fatal(err)
	}
	if hints := ValidateFeature("f", nil); len(hints) != 0 {
		t.Fatalf("the steward's verdict pair is not a chain hop, got %+v", hints)
	}
}
