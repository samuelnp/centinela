package main

import (
	"strings"
	"testing"
)

// mustRefuse runs `route set` and returns the refusal message.
func mustRefuse(t *testing.T, args ...string) string {
	t.Helper()
	err := runRouteSet(nil, args)
	if err == nil {
		t.Fatalf("expected %v to be refused", args)
	}
	if _, ok := routeOf(t, args[1]); ok {
		t.Fatalf("a refused route must persist nothing for %q", args[1])
	}
	return err.Error()
}

func TestRunRouteSet_StaticModeRefusesNamingTheKey(t *testing.T) {
	routeRepo(t, "plan", "")
	routeSetReason = "x"
	msg := mustRefuse(t, "f", "qa-senior", "fast")
	if !strings.Contains(msg, "routing_mode") || !strings.Contains(msg, "[orchestration]") {
		t.Fatalf("static-mode refusal must name the config key, got %q", msg)
	}
	if err := runRouteShow(nil, []string{"f"}); err == nil {
		t.Fatal("route show must refuse under static mode too")
	}
}

func TestRunRouteSet_FloorAndVocabularyRefusals(t *testing.T) {
	routeRepo(t, "plan", dynamicToml)
	routeSetReason = "save cost"
	if msg := mustRefuse(t, "f", "gatekeeper", "balanced"); !strings.Contains(msg, "reasoning") ||
		!strings.Contains(msg, "gatekeeper") {
		t.Fatalf("floor refusal must name the role and its floor, got %q", msg)
	}
	if msg := mustRefuse(t, "f", "wizard", "fast"); !strings.Contains(msg, "senior-engineer") {
		t.Fatalf("unknown role must list the allowed slugs, got %q", msg)
	}
	if msg := mustRefuse(t, "f", "qa-senior", "turbo"); !strings.Contains(msg, "reasoning, balanced, fast") {
		t.Fatalf("unknown tier must list the tiers, got %q", msg)
	}
}

func TestRunRouteSet_ReasonAndSchedulingRefusals(t *testing.T) {
	routeRepo(t, "plan", dynamicToml)
	routeSetReason = ""
	if msg := mustRefuse(t, "f", "qa-senior", "fast"); !strings.Contains(msg, "--reason") {
		t.Fatalf("a downgrade below the static default must demand --reason, got %q", msg)
	}
	routeSetReason = "x"
	// The feature is not user-facing here, so ux-ui-specialist is scheduled by
	// no step and a route for it would be consumed by nothing.
	if msg := mustRefuse(t, "f", "ux-ui-specialist", "fast"); !strings.Contains(msg, "not scheduled") {
		t.Fatalf("an unscheduled role must be refused, got %q", msg)
	}
	if msg := mustRefuse(t, "f", "merge-steward", "fast"); !strings.Contains(msg, "not scheduled") {
		t.Fatalf("an out-of-band role must be refused, got %q", msg)
	}
}

func TestRunRouteSet_CompletedWorkflowAndMissingFeature(t *testing.T) {
	routeRepo(t, "done", dynamicToml)
	routeSetReason = "too late"
	if msg := mustRefuse(t, "f", "planner", "balanced"); !strings.Contains(msg, "already complete") {
		t.Fatalf("a done workflow must refuse routing, got %q", msg)
	}
	if err := runRouteSet(nil, []string{"ghost", "planner", "balanced"}); err == nil {
		t.Fatal("an unknown feature must surface the missing-workflow error")
	}
}
