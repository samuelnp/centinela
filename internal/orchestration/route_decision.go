package orchestration

import (
	"fmt"
	"strings"
)

// RouteRequest is everything ValidateRoute needs to judge one `route set`. It is
// resolved by the caller and judged here: the rules are pure, do no IO, and read
// nothing but these fields.
type RouteRequest struct {
	Role Role
	// NewTier is the requested tier; CurrentTier is the role's effective tier
	// today (an existing route, else the static default); StaticTier is the
	// configured/built-in default the `--reason` rule is measured against.
	NewTier     Tier
	CurrentTier Tier
	StaticTier  Tier
	Floor       Tier
	HasFloor    bool
	Reason      string
	// Step is the role's scheduled step, named in the underway refusal.
	Step         string
	Scheduled    bool
	StepUnderway bool
}

// ParseRouteTarget applies refusal rules 2–3 in order: a completed workflow
// refuses every route, then the role and tier vocabularies are checked. It
// returns the normalized pair so the caller never re-derives it.
func ParseRouteTarget(workflowDone bool, role, tier string) (Role, Tier, error) {
	if workflowDone {
		return "", "", fmt.Errorf("workflow already complete — routing decisions are closed")
	}
	normalizedRole, ok := NormalizeRole(role)
	if !ok {
		return "", "", fmt.Errorf("unknown role %q (allowed: %s)", role, strings.Join(AllowedRoleSlugs(), ", "))
	}
	normalizedTier, ok := NormalizeTier(tier)
	if !ok {
		return "", "", fmt.Errorf("unknown tier %q (allowed: %s)", tier, allowedTiersList())
	}
	return normalizedRole, normalizedTier, nil
}

// ValidateRoute applies refusal rules 4–7 in order: the role must be scheduled
// by this workflow's steps, the tier must clear the role's floor, a DOWNGRADE is
// refused once the role's step is underway or completed, and any tier below the
// static default demands an explicit reason. Upgrades always pass rule 6.
func ValidateRoute(req RouteRequest) error {
	if !req.Scheduled {
		return fmt.Errorf("role %q is not scheduled in this workflow's steps — nothing would consume the route", req.Role)
	}
	if req.HasFloor && TierBelow(req.NewTier, req.Floor) {
		return fmt.Errorf("tier %q is below the %s floor %q — raise the tier or lower the floor in [orchestration.floors]",
			req.NewTier, req.Role, req.Floor)
	}
	if TierBelow(req.NewTier, req.CurrentTier) && req.StepUnderway {
		return fmt.Errorf("cannot downgrade %q to %q: its %q step is already underway or completed",
			req.Role, req.NewTier, req.Step)
	}
	if TierBelow(req.NewTier, req.StaticTier) && strings.TrimSpace(req.Reason) == "" {
		return fmt.Errorf("routing %q to %q is a downgrade below the static default %q — rerun with --reason \"…\"",
			req.Role, req.NewTier, req.StaticTier)
	}
	return nil
}

// NormalizeRole trims and lowercases s, then validates it against the role slugs
// config and routing may key on. ok is false for any unrecognized value.
func NormalizeRole(s string) (Role, bool) {
	candidate := strings.ToLower(strings.TrimSpace(s))
	for _, slug := range AllowedRoleSlugs() {
		if candidate == slug {
			return Role(slug), true
		}
	}
	return "", false
}

// allowedTiersList renders the tiers for an error message, in stable order.
func allowedTiersList() string {
	names := make([]string, 0, len(AllowedTiers()))
	for _, tier := range AllowedTiers() {
		names = append(names, string(tier))
	}
	return strings.Join(names, ", ")
}
