package acceptance_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/evidence"
	"github.com/samuelnp/centinela/internal/orchestration"
)

// Acceptance: specs/deterministic-artifact-scaffolds.feature
// Scenario: Init leaves outputs empty for every role
func TestDAS_InitOutputsEmptyEveryRole(t *testing.T) {
	roles := []evidence.Role{
		orchestration.RoleBigThinker, orchestration.RoleFeatureSpecial,
		orchestration.RoleSeniorEngineer,
	}
	for _, role := range roles {
		dasChdir(t, "demo.md")
		dasInit(t, "demo", role)
		_, e := dasReadJSON(t, "demo", role)
		if len(e.Outputs) != 0 {
			t.Fatalf("%s outputs not empty: %v", role, e.Outputs)
		}
	}
}

// Acceptance: specs/deterministic-artifact-scaffolds.feature
// Scenario: Init leaves edgeCases empty for every role
func TestDAS_InitEdgeCasesEmpty(t *testing.T) {
	dasChdir(t, "demo.md")
	dasInit(t, "demo", orchestration.RoleBigThinker)
	_, e := dasReadJSON(t, "demo", orchestration.RoleBigThinker)
	if len(e.EdgeCases) != 0 {
		t.Fatalf("edgeCases not empty: %v", e.EdgeCases)
	}
}

// Acceptance: specs/deterministic-artifact-scaffolds.feature
// Scenario: Pre-existing minimal evidence JSON still validates
func TestDAS_PreExistingMinimalJSONValidates(t *testing.T) {
	dasChdir(t, "a.md", "demo.md")
	if err := os.MkdirAll("docs/plans", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("docs/plans/demo.md", []byte("plan"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A hand-written minimal big-thinker evidence with manually-listed snapshot
	// inputs (no init pre-fill) must still satisfy the validator — no schema
	// change required.
	now := time.Now().UTC().Format(time.RFC3339)
	json := `{"feature":"demo","step":"plan","role":"big-thinker","status":"done",` +
		`"generatedAt":"` + now + `",` +
		`"inputs":["docs/features/a.md","docs/features/demo.md","docs/plans/demo.md"],` +
		`"outputs":["docs/plans/demo.md"],"edgeCases":[],"handoffTo":"feature-specialist"}`
	path := orchestration.JSONPath("demo", orchestration.RoleBigThinker)
	if err := os.WriteFile(path, []byte(json), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := orchestration.ValidateEvidence(path, "demo", "plan", orchestration.RoleBigThinker, nil); err != nil {
		if strings.Contains(err.Error(), "snapshot") {
			t.Fatalf("minimal hand-written JSON failed snapshot rule: %v", err)
		}
		t.Fatalf("minimal hand-written JSON failed validation: %v", err)
	}
}
