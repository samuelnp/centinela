// Acceptance: specs/dynamic-model-routing.feature
package acceptance_test

import (
	"strings"
	"testing"
)

// Scenario: Static mode output is byte-identical and route commands are refused
func TestDMR_StaticModeIsByteIdenticalAndRouteCommandsAreRefused(t *testing.T) {
	bin := dmrBuildBin(t)
	absentDir := dmrRepo(t, "")
	dmrWrite(t, absentDir, "f", dmrWorkflowJSON("f", "plan", ""))
	explicitDir := dmrRepo(t, dmrStaticTOML)
	dmrWrite(t, explicitDir, "f", dmrWorkflowJSON("f", "plan", ""))

	absent := dmrOK(t, bin, absentDir, "hook", "orchestration")
	explicit := dmrOK(t, bin, explicitDir, "hook", "orchestration")
	if absent != explicit {
		t.Fatalf("static mode drifted:\nabsent:\n%s\nexplicit:\n%s", absent, explicit)
	}
	for _, out := range []string{absent, explicit} {
		if strings.Contains(out, "routing") {
			t.Fatalf("static output must carry no routing directive: %s", out)
		}
	}
	refusal := dmrRefused(t, bin, explicitDir, "route", "set", "f", "qa-senior", "fast", "--reason", "x")
	for _, want := range []string{"routing_mode", "dynamic"} {
		mustContain(t, refusal, want)
	}
	mustContain(t, dmrRefused(t, bin, explicitDir, "route", "show", "f"), "routing_mode")
}

// Scenario: Static mode output is byte-identical and route commands are refused
//
// The case that would have caught E1's blast radius: byte-identity must hold
// with ROUTES ALREADY IN STATE, not only with an empty state. Static mode must
// ignore recorded routes entirely rather than read them and then decline.
func TestDMR_StaticModeIsByteIdenticalWithRoutesPresentInState(t *testing.T) {
	bin := dmrBuildBin(t)
	routes := `,"modelRoutes":{"senior-engineer":{"tier":"fast","reason":"x","decidedAt":"2026-01-02T00:00:00Z"},` +
		`"gatekeeper":{"tier":"fast","reason":"x","decidedAt":"2026-01-02T00:00:00Z"}}`
	clean := dmrRepo(t, dmrStaticTOML)
	dmrWrite(t, clean, "f", dmrWorkflowJSON("f", "code", ""))
	routed := dmrRepo(t, dmrStaticTOML)
	dmrWrite(t, routed, "f", dmrWorkflowJSON("f", "code", routes))
	absentKey := dmrRepo(t, "")
	dmrWrite(t, absentKey, "f", dmrWorkflowJSON("f", "code", routes))

	base := dmrOK(t, bin, clean, "hook", "orchestration")
	withRoutes := dmrOK(t, bin, routed, "hook", "orchestration")
	absent := dmrOK(t, bin, absentKey, "hook", "orchestration")
	if base != withRoutes || base != absent {
		t.Fatalf("routes in state must not move a static-mode byte:\n%s\n---\n%s\n---\n%s", base, withRoutes, absent)
	}
	mustContain(t, base, "senior-engineer (model: opus (claude)")
	if strings.Contains(base, "haiku") {
		t.Fatalf("static mode must not honor a recorded route: %s", base)
	}
}

// Scenario: A pre-existing workflow JSON without modelRoutes loads cleanly with no routes
func TestDMR_PreExistingWorkflowWithoutModelRoutesLoadsCleanly(t *testing.T) {
	bin, dir := dmrBuildBin(t), dmrRepo(t, dmrDynamicTOML)
	legacy := `{"feature":"old","startedAt":"` + dmrStartedAt + `","currentStep":"code",` +
		`"stepOrder":["plan","code","tests","validate","docs"],"steps":{},` +
		`"orchestrationMode":"strict-subagents-v1","planContract":"planner-v1",` +
		`"validateContract":"adversarial-v1"}`
	dmrWrite(t, dir, "old", legacy)

	show := dmrOK(t, bin, dir, "route", "show", "old")
	table, _, _ := strings.Cut(show, "\n\n") // the hint legitimately says "unrouted"
	if strings.Contains(table, "routed") || strings.Contains(table, "ignored") {
		t.Fatalf("a workflow with no routes must show every role as static: %s", show)
	}
	// The hooks behave exactly as before: identical bytes to static mode.
	staticDir := dmrRepo(t, dmrStaticTOML)
	dmrWrite(t, staticDir, "old", legacy)
	dynamicOut := dmrOK(t, bin, dir, "hook", "orchestration")
	staticOut := dmrOK(t, bin, staticDir, "hook", "orchestration")
	if !strings.Contains(dynamicOut, staticOut) {
		t.Fatalf("an un-routed dynamic workflow must keep the static directive verbatim:\n%s\n---\n%s", dynamicOut, staticOut)
	}
	mustContain(t, dmrOK(t, bin, dir, "status", "old"), "old")
	// Recording a route on the legacy state adds the key without migration.
	dmrOK(t, bin, dir, "route", "set", "old", "qa-senior", "fast", "--reason", "trivial")
	mustContain(t, dmrState(t, dir, "old"), `"modelRoutes"`)
}
