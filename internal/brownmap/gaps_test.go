package brownmap

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/reconstruct"
)

func TestGapPhases_TodosAndGoals(t *testing.T) {
	todos := []reconstruct.Target{{Pkg: "internal/svc", Slug: "svc", Role: reconstruct.RoleModule}}
	phases := gapPhases(todos, []string{"Add OAuth login"})
	if len(phases) != 1 {
		t.Fatalf("expected one gap phase, got %d", len(phases))
	}
	g := phases[0]
	if g.Name != GapPhaseName || len(g.Features) != 2 {
		t.Fatalf("gap phase mismatch: %q %d features", g.Name, len(g.Features))
	}
	if g.Features[0].Name != "svc-confirm" {
		t.Fatalf("TODO target must become a -confirm gap feature: %q", g.Features[0].Name)
	}
	if !strings.Contains(g.Features[0].Fixes, "unconfirmed") {
		t.Fatalf("TODO gap must carry a Fixes rationale: %q", g.Features[0].Fixes)
	}
	if g.Features[1].Name != "Add OAuth login" || g.Features[1].Source.Role != "operator-goal" {
		t.Fatalf("goal gap feature mismatch: %+v", g.Features[1])
	}
}

func TestGapPhases_NoWorkReturnsNil(t *testing.T) {
	if gapPhases(nil, nil) != nil {
		t.Fatal("no TODOs and no goals must yield nil gap phases")
	}
}

func TestGapPhases_GoalOnly(t *testing.T) {
	phases := gapPhases(nil, []string{"only goal"})
	if len(phases) != 1 || len(phases[0].Features) != 1 {
		t.Fatalf("goal-only must produce a single gap feature: %+v", phases)
	}
}
