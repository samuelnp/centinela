package acceptance_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/evidence"
	"github.com/samuelnp/centinela/internal/orchestration"
)

// Acceptance: specs/deterministic-artifact-scaffolds.feature
// Scenario: FillSlot renders the canonical marker
func TestDAS_FillSlotCanonicalMarker(t *testing.T) {
	if got := evidence.FillSlot("the impl file path"); got != "<FILL: the impl file path>" {
		t.Fatalf("FillSlot = %q", got)
	}
}

// Acceptance: specs/deterministic-artifact-scaffolds.feature
// Scenario: Companion skeleton seeds role-appropriate FILL slots
func TestDAS_CompanionSeedsRoleSlots(t *testing.T) {
	cases := []struct {
		role   evidence.Role
		header string
	}{
		{orchestration.RoleBigThinker, "Problem"},
		{orchestration.RoleFeatureSpecial, "Acceptance Criteria"},
		{orchestration.RoleSeniorEngineer, "Files Touched"},
		{orchestration.RoleQASeniorEngineer, "Test Inventory"},
		{orchestration.RoleValidationSpec, "Gates Run"},
		{orchestration.RoleUXUISpecialist, "Flow Review"},
		{orchestration.RoleDocsSpecialist, "KB Pages"},
	}
	for _, c := range cases {
		dasChdir(t, "demo.md")
		dasInit(t, "demo", c.role)
		body, err := os.ReadFile(".workflow/demo-" + string(c.role) + ".md")
		if err != nil {
			t.Fatal(err)
		}
		if !hasFill(string(body)) {
			t.Fatalf("%s companion missing fill marker", c.role)
		}
		if !strings.Contains(string(body), c.header) {
			t.Fatalf("%s companion missing header %q", c.role, c.header)
		}
	}
}

// Acceptance: specs/deterministic-artifact-scaffolds.feature
// Scenario: Unknown role falls back to the one-line companion placeholder
func TestDAS_UnknownRoleCompanionFallback(t *testing.T) {
	body := evidence.DefaultCompanionTemplate("demo", evidence.Role("merge-steward"))
	if hasFill(body) {
		t.Fatalf("unknown-role fallback must carry no fill marker: %s", body)
	}
	if !strings.Contains(body, "Replace this with the role's narrative report.") {
		t.Fatalf("unknown-role fallback missing one-line placeholder: %s", body)
	}
}

// Acceptance: specs/deterministic-artifact-scaffolds.feature
// Scenario: No FILL marker ever lands in an evidence JSON list field
func TestDAS_NoFillMarkerInJSONListField(t *testing.T) {
	dasChdir(t, "demo.md")
	dasInit(t, "demo", orchestration.RoleBigThinker)
	raw, e := dasReadJSON(t, "demo", orchestration.RoleBigThinker)
	if hasFill(raw) {
		t.Fatalf("fill marker leaked into JSON: %s", raw)
	}
	for _, lst := range [][]string{e.Inputs, e.Outputs, e.EdgeCases} {
		for _, v := range lst {
			if hasFill(v) {
				t.Fatalf("fill marker in list entry %q", v)
			}
		}
	}
}
