package evidence

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

const fillTok = "<FILL:"

func TestFillSlotRendersCanonicalMarker(t *testing.T) {
	if got := FillSlot("the impl file path"); got != "<FILL: the impl file path>" {
		t.Fatalf("FillSlot = %q", got)
	}
	if !strings.HasPrefix(FillMarker, fillTok) {
		t.Fatalf("FillMarker not the canonical template: %q", FillMarker)
	}
}

func TestCompanionSkeletonPerRoleHeaders(t *testing.T) {
	cases := map[Role]string{
		orchestration.RoleBigThinker:       "Problem",
		orchestration.RoleFeatureSpecial:   "Acceptance Criteria",
		orchestration.RoleSeniorEngineer:   "Files Touched",
		orchestration.RoleQASeniorEngineer: "Test Inventory",
		orchestration.RoleValidationSpec:   "Gates Run",
		orchestration.RoleUXUISpecialist:   "Flow Review",
		orchestration.RoleDocsSpecialist:   "KB Pages",
		Role("gatekeeper"):                 "Analyzed Specs",
	}
	for role, header := range cases {
		body, ok := companionSkeleton("demo", role)
		if !ok {
			t.Fatalf("%s: expected a skeleton", role)
		}
		if !strings.Contains(body, fillTok) {
			t.Fatalf("%s: skeleton missing fill marker", role)
		}
		if !strings.Contains(body, header) {
			t.Fatalf("%s: skeleton missing header %q", role, header)
		}
	}
}

func TestCompanionSkeletonUnknownRole(t *testing.T) {
	body, ok := companionSkeleton("demo", Role("merge-steward"))
	if ok || body != "" {
		t.Fatalf("unknown role should yield (\"\", false), got (%q, %v)", body, ok)
	}
}

func TestDefaultCompanionTemplateRoleAwareAndFallback(t *testing.T) {
	known := DefaultCompanionTemplate("demo", orchestration.RoleBigThinker)
	if !strings.Contains(known, "## Problem") || !strings.Contains(known, fillTok) {
		t.Fatalf("known-role template missing role section/fill: %s", known)
	}
	fallback := DefaultCompanionTemplate("demo", Role("merge-steward"))
	if strings.Contains(fallback, fillTok) {
		t.Fatalf("fallback should carry no fill marker: %s", fallback)
	}
	if !strings.Contains(fallback, "Replace this with the role's narrative report.") {
		t.Fatalf("fallback missing the legacy one-liner: %s", fallback)
	}
}

func TestSkeletonNeverEmitsFillMarkerInJSON(t *testing.T) {
	skel := Skeleton("demo", orchestration.RoleBigThinker, "1.0.0")
	skel.Inputs = PlanInputsStub()
	data, err := skel.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), fillTok) {
		t.Fatalf("fill marker leaked into JSON: %s", data)
	}
}

// PlanInputsStub returns a plausible pre-fill so the marshaled JSON exercises a
// populated inputs list (real paths, never a fill marker).
func PlanInputsStub() []string {
	return []string{"docs/features/demo.md", "docs/plans/demo.md"}
}
