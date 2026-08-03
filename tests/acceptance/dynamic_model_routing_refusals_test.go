// Acceptance: specs/dynamic-model-routing.feature
package acceptance_test

import (
	"strings"
	"testing"
)

// Scenario: A tier below the role's floor is refused naming the floor
func TestDMR_TierBelowTheRolesFloorIsRefusedNamingTheFloor(t *testing.T) {
	bin, dir := dmrBuildBin(t), dmrRepo(t, dmrDynamicTOML)
	dmrWrite(t, dir, "f", dmrWorkflowJSON("f", "plan", ""))

	out := dmrRefused(t, bin, dir, "route", "set", "f", "gatekeeper", "balanced", "--reason", "save cost")
	for _, want := range []string{"gatekeeper", "reasoning", "floor"} {
		mustContain(t, out, want)
	}
	if strings.Contains(dmrState(t, dir, "f"), "gatekeeper") {
		t.Fatalf("a refused route must leave no record: %s", dmrState(t, dir, "f"))
	}
}

// Scenario: A downgrade below the static default without --reason is refused
func TestDMR_DowngradeBelowTheStaticDefaultWithoutReasonIsRefused(t *testing.T) {
	bin, dir := dmrBuildBin(t), dmrRepo(t, dmrDynamicTOML)
	dmrWrite(t, dir, "f", dmrWorkflowJSON("f", "plan", ""))

	out := dmrRefused(t, bin, dir, "route", "set", "f", "qa-senior", "fast")
	for _, want := range []string{"balanced", "--reason"} {
		mustContain(t, out, want)
	}
	if strings.Contains(dmrState(t, dir, "f"), "qa-senior") {
		t.Fatalf("a refused route must leave no record: %s", dmrState(t, dir, "f"))
	}
	// Whitespace is not a reason.
	dmrRefused(t, bin, dir, "route", "set", "f", "qa-senior", "fast", "--reason", "   ")
}

// Scenario: A downgrade is refused once the role's step is underway
func TestDMR_DowngradeIsRefusedOnceTheRolesStepIsUnderway(t *testing.T) {
	bin, dir := dmrBuildBin(t), dmrRepo(t, dmrDynamicTOML)
	dmrWrite(t, dir, "f", dmrWorkflowJSON("f", "code", ""))
	dmrEvidence(t, dir, "f", "senior-engineer")

	out := dmrRefused(t, bin, dir, "route", "set", "f", "senior-engineer", "fast", "--reason", "late savings")
	for _, want := range []string{"senior-engineer", "code", "underway"} {
		mustContain(t, out, want)
	}
}

// Scenario: A downgrade is refused for a completed step
func TestDMR_DowngradeIsRefusedForACompletedStep(t *testing.T) {
	bin, dir := dmrBuildBin(t), dmrRepo(t, dmrDynamicTOML)
	dmrWrite(t, dir, "f", dmrWorkflowJSON("f", "code", ""))

	out := dmrRefused(t, bin, dir, "route", "set", "f", "planner", "balanced", "--reason", "too late")
	for _, want := range []string{"planner", "plan", "underway or completed"} {
		mustContain(t, out, want)
	}
}

// Scenario: Unknown role and unknown tier are refused
func TestDMR_UnknownRoleAndUnknownTierAreRefused(t *testing.T) {
	bin, dir := dmrBuildBin(t), dmrRepo(t, dmrDynamicTOML)
	dmrWrite(t, dir, "f", dmrWorkflowJSON("f", "plan", ""))

	role := dmrRefused(t, bin, dir, "route", "set", "f", "wizard", "fast", "--reason", "x")
	for _, want := range []string{"wizard", "planner", "senior-engineer", "gatekeeper"} {
		mustContain(t, role, want)
	}
	tier := dmrRefused(t, bin, dir, "route", "set", "f", "qa-senior", "turbo", "--reason", "x")
	mustContain(t, tier, "reasoning, balanced, fast")
}

// Scenario: A role not scheduled in this workflow's steps is refused
func TestDMR_RoleNotScheduledInThisWorkflowsStepsIsRefused(t *testing.T) {
	bin, dir := dmrBuildBin(t), dmrRepo(t, dmrDynamicTOML)
	dmrWrite(t, dir, "f", dmrWorkflowJSON("f", "plan", ""))

	// No docs/features/f.md ⇒ not user-facing ⇒ ux-ui-specialist is unscheduled.
	out := dmrRefused(t, bin, dir, "route", "set", "f", "ux-ui-specialist", "fast", "--reason", "x")
	mustContain(t, out, "not scheduled")
}

// Scenario: A role not scheduled in this workflow's steps is refused
//
// The archetype half: a hotfix workflow's step subset scopes the role set, so a
// role from a step this workflow never runs is refused rather than recorded.
func TestDMR_HotfixArchetypeScopesTheSchedulableRoles(t *testing.T) {
	bin, dir := dmrBuildBin(t), dmrRepo(t, dmrDynamicTOML)
	hotfix := `{"feature":"h","startedAt":"` + dmrStartedAt + `","currentStep":"code",` +
		`"stepOrder":["code","validate"],"steps":{},"orchestrationMode":"strict-subagents-v1",` +
		`"planContract":"planner-v1","validateContract":"adversarial-v1"}`
	dmrWrite(t, dir, "h", hotfix)

	for _, role := range []string{"planner", "qa-senior"} {
		mustContain(t, dmrRefused(t, bin, dir, "route", "set", "h", role, "fast", "--reason", "x"), "not scheduled")
	}
	dmrOK(t, bin, dir, "route", "set", "h", "senior-engineer", "balanced", "--reason", "small fix")
	show := dmrOK(t, bin, dir, "route", "show", "h")
	if strings.Contains(show, "qa-senior") {
		t.Fatalf("route show must scope to the archetype's steps: %s", show)
	}
}
