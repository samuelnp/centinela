package unit_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/analyze"
	"github.com/samuelnp/centinela/internal/brownmap"
	"github.com/samuelnp/centinela/internal/roadmap"
)

// brownInv is a Go n-tier fixture promoting behavioral packages.
func brownInv() analyze.Inventory {
	return analyze.Inventory{
		SchemaVersion: analyze.SchemaVersion, PrimaryLanguage: "Go",
		Packages: []string{"cmd/app", "internal/handler", "internal/service"},
		Graph:    analyze.DependencyGraph{Kind: "go-packages"},
	}
}

// TestBrownfield_BaselineExcludedFromStatusAndCoverage asserts the generated
// draft's Baseline phase is exempt from both the status counts and the validate
// coverage set, via the same predicate that already exempts Backlog.
func TestBrownfield_BaselineExcludedFromStatusAndCoverage(t *testing.T) {
	plan := brownmap.NewBrownfielder().Generate(brownInv(), nil)
	r := plan.Roadmap
	if !roadmap.IsBaselinePhaseName(r.Phases[0].Name) {
		t.Fatalf("first phase must use the Baseline convention, got %q", r.Phases[0].Name)
	}
	// Status counts only the schedulable gap features; the Baseline phase's
	// already-built features contribute nothing — exactly the Backlog exemption.
	planned, inProgress, done := r.Summary()
	if planned+inProgress+done != plan.GapCount {
		t.Fatalf("status must count only the %d gap features, got %d/%d/%d",
			plan.GapCount, planned, inProgress, done)
	}
	set := roadmap.NonBacklogFeatureSet(&r)
	for _, f := range r.Phases[0].Features {
		if set[f.Name] {
			t.Fatalf("Baseline feature %q must be excluded from the coverage set", f.Name)
		}
	}
}
