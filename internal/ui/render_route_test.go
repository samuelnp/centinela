package ui

import (
	"strings"
	"testing"
)

func TestRenderRouteTable_AlignsColumnsAndDashesEmptyCells(t *testing.T) {
	rows := []RouteRow{
		{Role: "planner", Tier: "reasoning", Source: "static", Floor: "balanced"},
		{Role: "senior-engineer", Tier: "balanced", Source: "routed", Reason: "trivial", DecidedAt: "2026-07-30T00:00:00Z"},
	}
	out := RenderRouteTable("f", rows, "")
	if !strings.Contains(out, "Model routing — f") {
		t.Fatalf("missing the table title: %s", out)
	}
	for _, want := range []string{"Role", "Tier", "Source", "Floor", "Reason", "Decided", "—"} {
		if !strings.Contains(out, want) {
			t.Fatalf("table missing %q: %s", want, out)
		}
	}
	header, first := strings.Split(out, "\n")[1], strings.Split(out, "\n")[2]
	if strings.Index(header, "Tier") != strings.Index(first, "reasoning") {
		t.Fatalf("columns are not aligned:\n%q\n%q", header, first)
	}
}

func TestRenderRouteTable_EmptyAndHint(t *testing.T) {
	if out := RenderRouteTable("f", nil, "ignored"); !strings.Contains(out, "no roles scheduled for f") {
		t.Fatalf("an empty table must render one muted line, got %q", out)
	}
	out := RenderRouteTable("f", []RouteRow{{Role: "planner", Tier: "fast", Source: "static"}}, "routing (dynamic): unrouted [planner]")
	if !strings.Contains(out, "routing (dynamic): unrouted [planner]") {
		t.Fatalf("the hint must be appended below the table: %s", out)
	}
}
