package orchestration

import (
	"strings"
	"testing"
)

func TestParseRouteTarget(t *testing.T) {
	if _, _, err := ParseRouteTarget(true, "qa-senior", "fast"); err == nil ||
		!strings.Contains(err.Error(), "already complete") {
		t.Fatalf("a done workflow must refuse routing, got %v", err)
	}
	_, _, err := ParseRouteTarget(false, "wizard", "fast")
	if err == nil || !strings.Contains(err.Error(), "senior-engineer") {
		t.Fatalf("an unknown role must list the allowed slugs, got %v", err)
	}
	_, _, err = ParseRouteTarget(false, "qa-senior", "turbo")
	if err == nil || !strings.Contains(err.Error(), "reasoning, balanced, fast") {
		t.Fatalf("an unknown tier must list the tiers, got %v", err)
	}
	role, tier, err := ParseRouteTarget(false, " QA-Senior ", " Fast ")
	if err != nil || role != RoleQASeniorEngineer || tier != TierFast {
		t.Fatalf("valid input must normalize, got %q/%q/%v", role, tier, err)
	}
}

// baseRequest is an accepted downgrade-with-reason before the step starts.
func baseRequest() RouteRequest {
	return RouteRequest{
		Role: RoleSeniorEngineer, NewTier: TierBalanced, CurrentTier: TierReasoning,
		StaticTier: TierReasoning, Reason: "config-only change", Step: "code", Scheduled: true,
	}
}

func TestValidateRoute_RefusalMatrix(t *testing.T) {
	if err := ValidateRoute(baseRequest()); err != nil {
		t.Fatalf("a reasoned pre-step downgrade must pass: %v", err)
	}
	unscheduled := baseRequest()
	unscheduled.Scheduled = false
	if err := ValidateRoute(unscheduled); err == nil || !strings.Contains(err.Error(), "not scheduled") {
		t.Fatalf("rule 4 must refuse an unscheduled role, got %v", err)
	}
	floored := baseRequest()
	floored.Role, floored.HasFloor, floored.Floor = RoleGatekeeper, true, TierReasoning
	err := ValidateRoute(floored)
	if err == nil || !strings.Contains(err.Error(), "gatekeeper") || !strings.Contains(err.Error(), "reasoning") {
		t.Fatalf("rule 5 must name the role and its floor, got %v", err)
	}
	underway := baseRequest()
	underway.StepUnderway = true
	err = ValidateRoute(underway)
	if err == nil || !strings.Contains(err.Error(), "code") || !strings.Contains(err.Error(), "underway") {
		t.Fatalf("rule 6 must name the underway step, got %v", err)
	}
	noReason := baseRequest()
	noReason.Reason = "   "
	if err := ValidateRoute(noReason); err == nil || !strings.Contains(err.Error(), "--reason") {
		t.Fatalf("rule 7 must demand a reason, got %v", err)
	}
}

func TestValidateRoute_UpgradesAlwaysPass(t *testing.T) {
	upgrade := baseRequest()
	upgrade.NewTier, upgrade.CurrentTier, upgrade.StaticTier = TierReasoning, TierBalanced, TierBalanced
	upgrade.StepUnderway, upgrade.Reason = true, ""
	if err := ValidateRoute(upgrade); err != nil {
		t.Fatalf("an upgrade mid-step needs no reason and is never refused: %v", err)
	}
}
