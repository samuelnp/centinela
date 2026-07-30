package main

import (
	"strings"
	"testing"
)

func TestRunRouteShow_EffectiveTableMixesRoutedAndStaticRows(t *testing.T) {
	routeRepo(t, "plan", dynamicToml)
	routeSetReason = "trivial rename"
	captureStdout(t, func() {
		if err := runRouteSet(nil, []string{"f", "qa-senior", "fast"}); err != nil {
			t.Fatalf("setup route: %v", err)
		}
	})
	routeSetReason = ""
	out := captureStdout(t, func() {
		if err := runRouteShow(nil, []string{"f"}); err != nil {
			t.Fatalf("show: %v", err)
		}
	})
	if !strings.Contains(out, "Role") || !strings.Contains(out, "Source") || !strings.Contains(out, "Decided") {
		t.Fatalf("table headers missing: %s", out)
	}
	for _, want := range []string{"qa-senior", "fast", "routed", "trivial rename"} {
		if !strings.Contains(out, want) {
			t.Fatalf("routed row missing %q: %s", want, out)
		}
	}
	if !strings.Contains(out, "gatekeeper") || !strings.Contains(out, "static") {
		t.Fatalf("un-routed roles must fall back to static: %s", out)
	}
	// Open question 4: the un-routed hint rides along while decisions are open.
	if !strings.Contains(out, "routing (dynamic): unrouted [planner]") {
		t.Fatalf("expected the routing hint for the current step: %s", out)
	}
}

// Compat: a workflow whose JSON predates modelRoutes shows every role as static.
func TestRunRouteShow_LegacyWorkflowIsAllStatic(t *testing.T) {
	routeRepo(t, "plan", dynamicToml)
	out := captureStdout(t, func() {
		if err := runRouteShow(nil, []string{"f"}); err != nil {
			t.Fatalf("show: %v", err)
		}
	})
	// The hint line legitimately says "unrouted", so only the table is checked.
	table, _, _ := strings.Cut(out, "\n\n")
	if strings.Contains(table, "routed") {
		t.Fatalf("a workflow with no routes must show no routed row: %s", out)
	}
	if !strings.Contains(out, "gatekeeper") || !strings.Contains(out, "reasoning") {
		t.Fatalf("the gatekeeper row must show its static tier and floor: %s", out)
	}
}
