package brownmap

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/reconstruct"
	"github.com/samuelnp/centinela/internal/roadmap"
)

func TestBaselinePhase_OneFeaturePerTargetInOrder(t *testing.T) {
	targets := []reconstruct.Target{
		{Pkg: "internal/a", Slug: "a", Role: reconstruct.RoleModule, Reason: "behavioral"},
		{Pkg: "cmd/b", Slug: "b", Role: reconstruct.RoleCommand, Reason: "cli surface"},
	}
	p := baselinePhase(targets)
	if p.Name != roadmap.BaselinePhaseName {
		t.Fatalf("phase name = %q", p.Name)
	}
	if len(p.Features) != 2 || p.Features[0].Name != "a" || p.Features[1].Name != "b" {
		t.Fatalf("baseline features mismatch: %+v", p.Features)
	}
	if p.Features[0].Source == nil || p.Features[0].Source.Feature != sourceFeature {
		t.Fatal("baseline feature must carry source provenance")
	}
	if !strings.Contains(p.Features[1].Description, "cmd/b") {
		t.Fatalf("description must reference the package: %q", p.Features[1].Description)
	}
}

func TestBaselinePhase_EmptyTargetsNonNilFeatures(t *testing.T) {
	p := baselinePhase(nil)
	if p.Features == nil {
		t.Fatal("empty Baseline phase must have non-nil Features slice")
	}
	if len(p.Features) != 0 {
		t.Fatalf("expected 0 features, got %d", len(p.Features))
	}
}

func TestRoleOrModule_NormalizesEmpty(t *testing.T) {
	if roleOrModule("") != reconstruct.RoleModule {
		t.Fatal("empty role must normalize to module")
	}
	if roleOrModule(reconstruct.RoleEndpoint) != reconstruct.RoleEndpoint {
		t.Fatal("known role must pass through unchanged")
	}
	d := baselineDescription(reconstruct.Target{Pkg: "p", Role: "", Reason: "r"})
	if !strings.Contains(d, string(reconstruct.RoleModule)) {
		t.Fatalf("empty-role description must render module label: %q", d)
	}
}
