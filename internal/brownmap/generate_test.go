package brownmap

import (
	"testing"

	"github.com/samuelnp/centinela/internal/analyze"
	"github.com/samuelnp/centinela/internal/roadmap"
)

// goInv is a Go n-tier fixture that promotes behavioral packages, so Generate
// yields a non-empty Baseline phase plus TODO-seeded gaps.
func goInv() analyze.Inventory {
	return analyze.Inventory{
		SchemaVersion: analyze.SchemaVersion, PrimaryLanguage: "Go",
		Packages: []string{"cmd/app", "internal/handler", "internal/service"},
		Graph:    analyze.DependencyGraph{Kind: "go-packages"},
	}
}

func TestGenerate_BaselinePlusGapsFromTodosAndGoal(t *testing.T) {
	p := NewBrownfielder().Generate(goInv(), []string{"Add OAuth login"})
	if p.Roadmap.Phases[0].Name != roadmap.BaselinePhaseName {
		t.Fatalf("first phase must be Baseline, got %q", p.Roadmap.Phases[0].Name)
	}
	if p.BaselineCount == 0 {
		t.Fatal("expected at least one Baseline feature")
	}
	gap := p.Roadmap.Phases[len(p.Roadmap.Phases)-1]
	if gap.Name != GapPhaseName {
		t.Fatalf("expected a Gaps phase, got %q", gap.Name)
	}
	var sawGoal bool
	for _, f := range gap.Features {
		if f.Name == "Add OAuth login" {
			sawGoal = true
		}
	}
	if !sawGoal {
		t.Fatal("goal-derived feature must live in the gap phase")
	}
	if p.GapCount == 0 || p.DraftPath != DefaultDraftPath {
		t.Fatalf("gapCount=%d draftPath=%q", p.GapCount, p.DraftPath)
	}
}

func TestGenerate_EmptyInventoryEmptyBaselineNoGaps(t *testing.T) {
	p := NewBrownfielder().Generate(analyze.Inventory{
		SchemaVersion: analyze.SchemaVersion, PrimaryLanguage: "Markdown",
		Packages: []string{"docs", "readme"},
	}, nil)
	if p.BaselineCount != 0 || p.GapCount != 0 {
		t.Fatalf("doc-only inventory must yield 0 baseline 0 gaps, got %d/%d", p.BaselineCount, p.GapCount)
	}
	if len(p.Roadmap.Phases) != 1 || p.Roadmap.Phases[0].Name != roadmap.BaselinePhaseName {
		t.Fatalf("empty inventory must keep a single Baseline phase, got %+v", p.Roadmap.Phases)
	}
	if p.Roadmap.Phases[0].Features == nil {
		t.Fatal("Baseline features must be non-nil for a stable draft")
	}
}

func TestGenerate_Deterministic(t *testing.T) {
	a := NewBrownfielder().Generate(goInv(), []string{"g1", "g2"})
	b := NewBrownfielder().Generate(goInv(), []string{"g1", "g2"})
	if a.BaselineCount != b.BaselineCount || a.GapCount != b.GapCount {
		t.Fatal("counts must match across runs")
	}
	for i := range a.Roadmap.Phases {
		for j := range a.Roadmap.Phases[i].Features {
			if a.Roadmap.Phases[i].Features[j].Name != b.Roadmap.Phases[i].Features[j].Name {
				t.Fatal("feature order must be byte-stable across runs")
			}
		}
	}
}
