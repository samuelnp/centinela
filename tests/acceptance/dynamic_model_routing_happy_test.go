// Acceptance: specs/dynamic-model-routing.feature
package acceptance_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var dmrRFC3339 = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)

// Scenario: Downgrade with a reason is recorded before the step starts
func TestDMR_DowngradeWithReasonIsRecordedBeforeTheStepStarts(t *testing.T) {
	bin, dir := dmrBuildBin(t), dmrRepo(t, dmrDynamicTOML)
	dmrWrite(t, dir, "f", dmrWorkflowJSON("f", "plan", ""))

	dmrOK(t, bin, dir, "route", "set", "f", "senior-engineer", "balanced", "--reason", "config-only change")

	state := dmrState(t, dir, "f")
	for _, want := range []string{`"modelRoutes"`, `"senior-engineer"`, `"tier": "balanced"`, `"reason": "config-only change"`} {
		mustContain(t, state, want)
	}
	decided := regexp.MustCompile(`"decidedAt": "([^"]+)"`).FindStringSubmatch(state)
	if len(decided) != 2 || !dmrRFC3339.MatchString(decided[1]) {
		t.Fatalf("decidedAt must be RFC3339 UTC, got %v", decided)
	}
	events, err := os.ReadFile(filepath.Join(dir, ".workflow", "telemetry", "events.jsonl"))
	if err != nil {
		t.Fatalf("telemetry not written: %v", err)
	}
	for _, want := range []string{`"route-decision"`, `"role":"senior-engineer"`, `"tier":"balanced"`,
		`"prevTier":"reasoning"`, `"reason":"config-only change"`} {
		mustContain(t, string(events), want)
	}
}

// Scenario: The hook directive reflects a routed tier and static config for un-routed roles
func TestDMR_HookDirectiveReflectsRoutedTierAndStaticConfig(t *testing.T) {
	bin, dir := dmrBuildBin(t), dmrRepo(t, dmrDynamicTOML)
	dmrWrite(t, dir, "f", dmrWorkflowJSON("f", "plan", ""))
	dmrOK(t, bin, dir, "route", "set", "f", "senior-engineer", "balanced", "--reason", "config-only change")
	dmrWrite(t, dir, "f", strings.Replace(dmrState(t, dir, "f"), `"currentStep": "plan"`, `"currentStep": "code"`, 1))

	out := dmrOK(t, bin, dir, "hook", "orchestration")
	mustContain(t, out, "senior-engineer (model: sonnet (claude)")
	// Un-routed roles keep resolving through the built-in defaults.
	ahead := dmrRepo(t, dmrDynamicTOML)
	dmrWrite(t, ahead, "f", dmrWorkflowJSON("f", "code", ""))
	mustContain(t, dmrOK(t, bin, ahead, "hook", "orchestration"), "senior-engineer (model: opus (claude)")
}

// Scenario: Upgrade is allowed anytime without a reason
func TestDMR_UpgradeIsAllowedAnytimeWithoutAReason(t *testing.T) {
	bin, dir := dmrBuildBin(t), dmrRepo(t, dmrDynamicTOML)
	routes := `,"modelRoutes":{"senior-engineer":{"tier":"balanced","reason":"earlier call","decidedAt":"2026-01-02T00:00:00Z"}}`
	dmrWrite(t, dir, "f", dmrWorkflowJSON("f", "code", routes))
	dmrEvidence(t, dir, "f", "senior-engineer") // the step is underway

	out := dmrOK(t, bin, dir, "route", "set", "f", "senior-engineer", "reasoning")
	mustContain(t, out, "senior-engineer → reasoning (was balanced)")
	state := dmrState(t, dir, "f")
	mustContain(t, state, `"tier": "reasoning"`)
	if strings.Contains(state, `"reason"`) {
		t.Fatalf("an upgrade with no --reason must record an empty reason: %s", state)
	}
}

// Scenario: route show renders the effective table with static fallbacks
func TestDMR_RouteShowRendersTheEffectiveTableWithStaticFallbacks(t *testing.T) {
	bin, dir := dmrBuildBin(t), dmrRepo(t, dmrDynamicTOML)
	dmrWrite(t, dir, "f", dmrWorkflowJSON("f", "plan", ""))
	dmrOK(t, bin, dir, "route", "set", "f", "qa-senior", "fast", "--reason", "trivial rename")

	out := dmrOK(t, bin, dir, "route", "show", "f")
	for _, role := range []string{"planner", "senior-engineer", "qa-senior", "gatekeeper"} {
		mustContain(t, out, role)
	}
	qa := dmrRow(t, out, "qa-senior")
	for _, want := range []string{"fast", "routed", "trivial rename", "2026-"} {
		mustContain(t, qa, want)
	}
	gate := dmrRow(t, out, "gatekeeper")
	for _, want := range []string{"static", "reasoning"} {
		mustContain(t, gate, want)
	}
}

// Scenario: The routing directive is emitted while roles are unset and silent once routed
func TestDMR_RoutingDirectiveIsEmittedWhileUnsetAndSilentOnceRouted(t *testing.T) {
	bin, dir := dmrBuildBin(t), dmrRepo(t, dmrDynamicTOML)
	dmrWrite(t, dir, "f", dmrWorkflowJSON("f", "plan", ""))

	out := dmrOK(t, bin, dir, "hook", "orchestration")
	for _, want := range []string{"routing (dynamic)", "unrouted [planner]", "planner>=balanced", "centinela route set f"} {
		mustContain(t, out, want)
	}
	dmrOK(t, bin, dir, "route", "set", "f", "planner", "reasoning")
	if after := dmrOK(t, bin, dir, "hook", "orchestration"); strings.Contains(after, "routing (dynamic)") {
		t.Fatalf("a fully routed step must emit no routing line: %s", after)
	}
}

// dmrRow returns the route-show table line whose first column is role.
func dmrRow(t *testing.T, out, role string) string {
	t.Helper()
	for _, line := range strings.Split(out, "\n") {
		if fields := strings.Fields(line); len(fields) > 0 && fields[0] == role {
			return line
		}
	}
	t.Fatalf("no table row for %q in:\n%s", role, out)
	return ""
}
