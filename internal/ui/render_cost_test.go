package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/cost"
)

func TestRenderCostEmpty(t *testing.T) {
	out := RenderCost(cost.Report{})
	if !strings.Contains(out, "no cost samples yet") {
		t.Fatalf("empty report should say so, got %q", out)
	}
}

func TestRenderCostRowsAndOverMarker(t *testing.T) {
	r := cost.Report{
		Features: []cost.Status{{Scope: "feature", Name: "f", Used: 100, Budget: 500}},
		Steps:    []cost.Status{{Scope: "step", Name: "f/code", Used: 1100, Budget: 1000, Over: true}},
		Models:   []cost.Status{{Scope: "model", Name: "m", Used: 50, Budget: 0}}, // no budget
	}
	out := RenderCost(r)
	if !strings.Contains(out, "f/code") || !strings.Contains(out, "OVER") {
		t.Fatalf("expected over-budget step row, got %q", out)
	}
	if !strings.Contains(out, "By model") {
		t.Fatalf("expected model section, got %q", out)
	}
}

func TestRenderCostWarningLine(t *testing.T) {
	s := cost.Status{Scope: "step", Name: "f/code", Used: 1100, Budget: 1000, Over: true}
	out := RenderCostWarning(s)
	if !strings.Contains(out, "over budget") || !strings.Contains(out, "f/code") {
		t.Fatalf("warning missing detail: %q", out)
	}
}

func TestRenderCostEmptySection(t *testing.T) {
	r := cost.Report{Features: []cost.Status{{Name: "f", Used: 1, Budget: 0}}}
	if !strings.Contains(RenderCost(r), "(none)") {
		t.Fatal("expected (none) for empty step/model sections")
	}
}
