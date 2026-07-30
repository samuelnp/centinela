package main

import (
	"strings"
	"testing"
)

// hookOut runs the orchestration hook in the current repo and returns stdout.
func hookOut(t *testing.T) string {
	t.Helper()
	return captureStdout(t, func() {
		withStdin(t, "{}", func() { runHookOrchestration(nil, nil) }) //nolint:errcheck
	})
}

// The headline invariant: an absent routing_mode and an explicit "static" must
// produce byte-identical output, and neither may grow a routing line.
func TestRunHookOrchestration_StaticModeIsByteIdentical(t *testing.T) {
	routeRepo(t, "plan", "")
	absent := hookOut(t)
	routeRepo(t, "plan", "[orchestration]\nrouting_mode = \"static\"\n")
	explicit := hookOut(t)
	if absent != explicit {
		t.Fatalf("static mode drifted:\nabsent:\n%s\nexplicit:\n%s", absent, explicit)
	}
	for _, out := range []string{absent, explicit} {
		if strings.Contains(out, "routing") {
			t.Fatalf("static output must carry no routing directive: %s", out)
		}
	}
}

func TestRunHookOrchestration_DynamicEmitsDirectiveUntilRouted(t *testing.T) {
	routeRepo(t, "plan", dynamicToml)
	out := hookOut(t)
	for _, want := range []string{"routing (dynamic): unrouted [planner]", "floors: planner>=balanced", "centinela route set f"} {
		if !strings.Contains(out, want) {
			t.Fatalf("dynamic directive missing %q: %s", want, out)
		}
	}
	routeSetReason = ""
	captureStdout(t, func() {
		if err := runRouteSet(nil, []string{"f", "planner", "reasoning"}); err != nil {
			t.Fatalf("route set: %v", err)
		}
	})
	if after := hookOut(t); strings.Contains(after, "routing (dynamic)") {
		t.Fatalf("a fully routed step must emit no routing line: %s", after)
	}
}

// A routed tier flows into the existing model: annotations; un-routed roles keep
// resolving through [orchestration.models] and the built-in defaults.
func TestRunHookOrchestration_DynamicOverlayShowsRoutedModel(t *testing.T) {
	routeRepo(t, "code", dynamicToml)
	routeSetReason = "config-only change"
	captureStdout(t, func() {
		if err := runRouteSet(nil, []string{"f", "senior-engineer", "balanced"}); err != nil {
			t.Fatalf("route set: %v", err)
		}
	})
	out := hookOut(t)
	if !strings.Contains(out, "senior-engineer (model: sonnet (claude)") {
		t.Fatalf("routed tier must reach the model annotation: %s", out)
	}
}
