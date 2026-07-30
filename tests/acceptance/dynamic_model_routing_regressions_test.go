// Acceptance: specs/dynamic-model-routing.feature
//
// Regression guards for the defects the edge-case tester reproduced against the
// shipped binary (E1-E4, E6-E8). Each drives the real binary, so each fails on
// the pre-fix code and passes on the fixed code.
package acceptance_test

import (
	"os"
	"strings"
	"testing"
	"time"
)

// Scenario: A tier below the role's floor is refused naming the floor
//
// E1 (CRITICAL) — `.workflow/<f>.json` is agent-writable in every step, so a
// floor enforced only on the `route set` write path is not a floor. The tier the
// hook RESOLVES must ignore a hand-written sub-floor route.
func TestDMR_HandWrittenGatekeeperRouteIsNotHonored(t *testing.T) {
	bin, dir := dmrBuildBin(t), dmrRepo(t, dmrDynamicTOML)
	attack := `,"modelRoutes":{"gatekeeper":{"tier":"fast","reason":"cost","decidedAt":"2026-01-02T00:00:00Z"}}`
	dmrWrite(t, dir, "g", dmrWorkflowJSON("g", "validate", attack))

	out := dmrOK(t, bin, dir, "hook", "orchestration")
	mustContain(t, out, "gatekeeper (model: opus (claude)")
	if strings.Contains(out, "haiku") {
		t.Fatalf("the trust anchor must not be downgradable by hand-written state: %s", out)
	}
	// The same value through the sanctioned path is refused, so the two paths
	// now agree — the write path and the resolution path enforce one rule.
	mustContain(t, dmrRefused(t, bin, dir, "route", "set", "g", "gatekeeper", "fast", "--reason", "cost"), "floor")
	// And the audit surface reports it as ignored rather than effective.
	row := dmrRow(t, dmrOK(t, bin, dir, "route", "show", "g"), "gatekeeper")
	for _, want := range []string{"reasoning", "ignored"} {
		mustContain(t, row, want)
	}
	// An ignored route leaves the decision OPEN, so the directive asks again.
	mustContain(t, out, "unrouted [gatekeeper]")
}

// Scenario: A tier below the role's floor is refused naming the floor
//
// E2 (HIGH) — a legacy plan contract schedules the retired plan roles; without
// the reverse alias they inherit no floor and the whole plan step is floorless.
func TestDMR_LegacyPlanRoleDowngradeIsRefused(t *testing.T) {
	bin, dir := dmrBuildBin(t), dmrRepo(t, dmrDynamicTOML)
	dmrWrite(t, dir, "lg", dmrLegacyWorkflowJSON("lg", "plan", ""))

	out := dmrRefused(t, bin, dir, "route", "set", "lg", "big-thinker", "fast", "--reason", "cheap plan")
	for _, want := range []string{"big-thinker", "balanced", "floor"} {
		mustContain(t, out, want)
	}
	// The validate half: a legacy contract schedules validation-specialist and
	// no gatekeeper row at all, so it must inherit the gatekeeper floor.
	mustContain(t, dmrRefused(t, bin, dir, "route", "set", "lg", "validation-specialist", "fast", "--reason", "x"), "floor")
	// And the resolution path holds the same line against hand-written state.
	attack := `,"modelRoutes":{"big-thinker":{"tier":"fast","reason":"x","decidedAt":"2026-01-02T00:00:00Z"}}`
	dmrWrite(t, dir, "lg", dmrLegacyWorkflowJSON("lg", "plan", attack))
	hook := dmrOK(t, bin, dir, "hook", "orchestration")
	mustContain(t, hook, "big-thinker (model: opus (claude)")
	if strings.Contains(hook, "haiku") {
		t.Fatalf("a legacy plan role must not be routable below the planner floor: %s", hook)
	}
}

// Scenario: The hook directive reflects a routed tier and static config for un-routed roles
//
// E4 (MED-HIGH) — the directive prints `static: senior-engineer=reasoning`, so
// confirming that very tier is the natural way to silence the un-routed line. It
// must not cost the project its per-runner concrete-model pin.
func TestDMR_SameTierRouteKeepsPinnedModel(t *testing.T) {
	pin := dmrDynamicTOML + "\n[orchestration.models]\nsenior-engineer = { claude = \"my-pinned-sonnet-4-9\" }\n"
	bin, dir := dmrBuildBin(t), dmrRepo(t, pin)
	dmrWrite(t, dir, "h", dmrWorkflowJSON("h", "code", ""))

	before := dmrOK(t, bin, dir, "hook", "orchestration")
	mustContain(t, before, "model: my-pinned-sonnet-4-9 (claude)")

	dmrOK(t, bin, dir, "route", "set", "h", "senior-engineer", "reasoning")
	after := dmrOK(t, bin, dir, "hook", "orchestration")
	mustContain(t, after, "model: my-pinned-sonnet-4-9 (claude)")
	if strings.Contains(after, "model: opus (claude)") {
		t.Fatalf("a same-tier route must not discard the pin: %s", after)
	}
	// A route to a DIFFERENT tier legitimately replaces the pinned entry.
	dmrOK(t, bin, dir, "route", "set", "h", "senior-engineer", "balanced", "--reason", "config-only change")
	moved := dmrOK(t, bin, dir, "hook", "orchestration")
	mustContain(t, moved, "model: sonnet (claude)")
	if strings.Contains(moved, "my-pinned-sonnet-4-9") {
		t.Fatalf("a real tier change must take effect: %s", moved)
	}
}

// Scenario: The routing directive is emitted while roles are unset and silent once routed
//
// E3 (HIGH) — floors bind ROUTES, never the project's explicit static config.
// Printing a bare "floors:" beside a lower "static:" claimed an enforcement the
// static path deliberately does not perform.
func TestDMR_DirectiveLabelsFloorsAsRouteOnly(t *testing.T) {
	toml := dmrDynamicTOML + "\n[orchestration.models]\ngatekeeper = \"fast\"\n"
	bin, dir := dmrBuildBin(t), dmrRepo(t, toml)
	dmrWrite(t, dir, "f", dmrWorkflowJSON("f", "validate", ""))

	out := dmrOK(t, bin, dir, "hook", "orchestration")
	mustContain(t, out, "floors (routes only): gatekeeper>=reasoning")
	// The static choice still stands — the fix is honesty, not new enforcement.
	mustContain(t, out, "static: gatekeeper=fast")
	mustContain(t, dmrOK(t, bin, dir, "route", "show", "f"), "Route floor")
}

// Scenario: Downgrade with a reason is recorded before the step starts
//
// E6 (MED) — a stub left over from an earlier run of the same slug must not
// close the sanctioned start-time routing window.
func TestDMR_StaleEvidenceDoesNotCloseTheRoutingWindow(t *testing.T) {
	bin, dir := dmrBuildBin(t), dmrRepo(t, dmrDynamicTOML)
	started := time.Now().UTC()
	body := `{"feature":"pr","startedAt":"` + started.Format(time.RFC3339) + `","currentStep":"code",` +
		`"stepOrder":["plan","code","tests","validate","docs"],"steps":{},` +
		`"orchestrationMode":"strict-subagents-v1","planContract":"planner-v1",` +
		`"validateContract":"adversarial-v1"}`
	dmrWrite(t, dir, "pr", body)
	stub := dmrEvidence(t, dir, "pr", "senior-engineer")
	stale := started.Add(-2 * time.Hour)
	if err := os.Chtimes(stub, stale, stale); err != nil {
		t.Fatal(err)
	}

	dmrOK(t, bin, dir, "route", "set", "pr", "senior-engineer", "balanced", "--reason", "config-only change")
	// Refreshed during this run, the same stub closes the window again.
	fresh := started.Add(time.Minute)
	if err := os.Chtimes(stub, fresh, fresh); err != nil {
		t.Fatal(err)
	}
	out := dmrRefused(t, bin, dir, "route", "set", "pr", "senior-engineer", "fast", "--reason", "late savings")
	mustContain(t, out, "underway")
}

// Scenario: A downgrade is refused once the role's step is underway
//
// E7 (MED) — a currentStep outside the step order used to disarm every underway
// refusal at once. The fail-safe reading is "underway".
func TestDMR_OutOfOrderCurrentStepStillRefusesDowngrades(t *testing.T) {
	bin, dir := dmrBuildBin(t), dmrRepo(t, dmrDynamicTOML)
	body := `{"feature":"k","startedAt":"` + dmrStartedAt + `","currentStep":"foo",` +
		`"stepOrder":["plan","code","tests","validate","docs"],"steps":{},` +
		`"orchestrationMode":"strict-subagents-v1","planContract":"planner-v1",` +
		`"validateContract":"adversarial-v1"}`
	dmrWrite(t, dir, "k", body)

	out := dmrRefused(t, bin, dir, "route", "set", "k", "senior-engineer", "fast", "--reason", "late savings")
	mustContain(t, out, "underway")
	// Upgrades are still never gated by rule 6.
	dmrOK(t, bin, dir, "route", "set", "k", "qa-senior", "reasoning")
}

// Scenario: route show renders the effective table with static fallbacks
//
// E8 (MED) — the audit surface and the effective behavior must agree: a corrupt
// recorded tier is ignored by the overlay, so it must not render as effective.
func TestDMR_RouteShowMatchesHookDirective(t *testing.T) {
	bin, dir := dmrBuildBin(t), dmrRepo(t, dmrDynamicTOML)
	corrupt := `,"modelRoutes":{"senior-engineer":{"tier":"ultra","reason":"x","decidedAt":"2026-01-02T00:00:00Z"}}`
	dmrWrite(t, dir, "c", dmrWorkflowJSON("c", "code", corrupt))

	show := dmrOK(t, bin, dir, "route", "show", "c")
	row := dmrRow(t, show, "senior-engineer")
	if strings.Contains(row, "ultra") {
		t.Fatalf("a corrupt tier must never render as the effective one: %s", row)
	}
	for _, want := range []string{"reasoning", "ignored"} {
		mustContain(t, row, want)
	}
	mustContain(t, dmrOK(t, bin, dir, "hook", "orchestration"), "senior-engineer (model: opus (claude)")
}
