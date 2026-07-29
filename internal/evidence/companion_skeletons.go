package evidence

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// companionHeaders maps a role to its LOCKED section headers, aligned to the
// live *-prompt.md docs. Markdown-only; the validator enforces existence, not
// these headers, so drift here is cosmetic.
var companionHeaders = map[Role][]string{
	// The planner's headers are the ordered union of both lenses: strategy
	// first, then spec, matching docs/architecture/planner-prompt.md.
	orchestration.RolePlanner: {"Problem", "Scope", "Dependencies & Assumptions", "Risks", "Rollout",
		"Behavior Summary", "Acceptance Criteria (Gherkin)", "UX States", "Edge Cases", "Out-of-Scope", "Handoff"},
	orchestration.RoleBigThinker:       {"Problem", "Scope", "Dependencies & Assumptions", "Risks", "Rollout", "Handoff"},
	orchestration.RoleFeatureSpecial:   {"Behavior Summary", "Acceptance Criteria (Gherkin)", "UX States", "Edge Cases", "Out-of-Scope", "Handoff"},
	orchestration.RoleSeniorEngineer:   {"Files Touched", "Architecture Compliance", "Type-Safety Notes", "Trade-Offs", "Handoff"},
	orchestration.RoleUXUISpecialist:   {"Flow Review", "Accessibility", "Visual Hierarchy", "State Coverage", "Handoff"},
	orchestration.RoleQASeniorEngineer: {"Test Inventory", "Coverage Gaps", "Acceptance Wiring", "Handoff"},
	orchestration.RoleValidationSpec:   {"Gates Run", "Synthesis", "Decision"},
	orchestration.RoleDocsSpecialist:   {"KB Pages", "project-docs Entries", "Outcome"},
	orchestration.RoleGatekeeper:       {"Inputs Read", "Analyzed Specs", "Refutation Attempts", "Commands Run", "Findings", "Recommendation"},
	Role("production-readiness"):       {"Files Reviewed", "Findings", "Recommendation"},
}

// companionSkeleton renders the per-role markdown skeleton seeded with FILL
// slots, returning (body, true) when the role has one. The header is added by
// the caller (DefaultCompanionTemplate).
func companionSkeleton(_ string, role Role) (string, bool) {
	headers, ok := companionHeaders[role]
	if !ok {
		return "", false
	}
	var b strings.Builder
	for _, h := range headers {
		fmt.Fprintf(&b, "## %s\n\n%s\n\n", h, FillSlot(strings.ToLower(h)))
	}
	return b.String(), true
}
