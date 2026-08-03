package evidence

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

// prefillRepo chdirs into a temp repo with a saved workflow, so the prefill
// has a real contract to derive from rather than the no-state fallback.
func prefillRepo(t *testing.T, feature string, userFacing bool) {
	t.Helper()
	t.Chdir(t.TempDir())
	if err := os.MkdirAll("docs/features", 0o755); err != nil {
		t.Fatal(err)
	}
	brief := "# " + feature + "\n"
	if userFacing {
		brief += "surface: user-facing\n"
	}
	if err := os.WriteFile(filepath.Join("docs/features", feature+".md"), []byte(brief), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(workflow.WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := workflow.Save(workflow.New(feature)); err != nil {
		t.Fatal(err)
	}
}

// The CLI's own prefill must never seed a value its own gate then refuses.
// On a user-facing feature the code step's successor is the SAME-step
// ux-ui-specialist, which the old static table got wrong for every such
// feature — it seeded qa-senior, a role one step further on.
func TestSkeletonPrefillDerivesFromTheWorkflowContract(t *testing.T) {
	prefillRepo(t, "uf", true)
	for role, want := range map[Role]string{
		orchestration.RoleSeniorEngineer:   "ux-ui-specialist",
		orchestration.RoleUXUISpecialist:   "qa-senior",
		orchestration.RoleQASeniorEngineer: "gatekeeper",
		orchestration.RoleGatekeeper:       "documentation-specialist",
		orchestration.RoleDocsSpecialist:   "complete",
	} {
		if got := Skeleton("uf", role, "0").HandoffTo; got != want {
			t.Errorf("%s handoffTo = %q, want %q", role, got, want)
		}
	}
}

// An internal feature's docs step requires no role evidence, so its validate
// step is terminal — the static table's "documentation-specialist" would have
// been a handoff to a role the workflow never asks for.
func TestSkeletonPrefillTerminatesOnInternalFeature(t *testing.T) {
	prefillRepo(t, "in", false)
	for role, want := range map[Role]string{
		orchestration.RoleSeniorEngineer:   "qa-senior",
		orchestration.RoleQASeniorEngineer: "gatekeeper",
		orchestration.RoleGatekeeper:       "complete",
		orchestration.RoleMergeSteward:     "complete",
	} {
		if got := Skeleton("in", role, "0").HandoffTo; got != want {
			t.Errorf("%s handoffTo = %q, want %q", role, got, want)
		}
	}
}

// The re-export is the only thing keeping the renderer and the gates on one
// string; if it ever drifts, a scaffold stops being detectable as a scaffold.
func TestFillMarkerAliasParity(t *testing.T) {
	if FillMarker != orchestration.FillMarker {
		t.Fatalf("FillMarker %q != orchestration.FillMarker %q", FillMarker, orchestration.FillMarker)
	}
	if FillSlot("x") != orchestration.FillSlot("x") {
		t.Fatal("FillSlot must delegate to the canonical renderer")
	}
	if !orchestration.UnreplacedSlot(string(changelogBody("f"))) {
		t.Fatal("the scaffolded changelog must be detectable as an unreplaced template")
	}
}
