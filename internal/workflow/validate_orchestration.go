package workflow

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
)

func validateOrchestration(feature, step string, cfg *config.Config) error {
	if !strictOrchestrationEnabled(feature) {
		return nil
	}
	roles := RequiredEvidenceRoles(feature, step)
	err := orchestration.ValidateRoles(feature, step, roles, config.UIPaths(cfg))
	if err == nil {
		// Only once the evidence is structurally sound: a chain complaint about
		// a file that is missing or schema-invalid is noise on top of a defect
		// ValidateRoles has already named.
		err = validateHandoffChain(feature, step, roles)
	}
	return annotatePlanContract(feature, step, err)
}

// RequiredEvidenceRoles resolves the evidence roles for a step. The validate
// and plan steps are the two places the pinned contract overrides policy: a
// workflow started before adversarial-v1 keeps demanding validation-specialist
// evidence, and one started before planner-v1 keeps demanding the complete
// big-thinker + feature-specialist pair. Each is what the workflow actually
// wrote, and neither is ever retro-gated on the newer role.
//
// EVERY caller that tells an operator what to produce MUST use this, not
// orchestration.RequiredRolesForFeature — the policy layer cannot see the
// pinned contract, so a contract-blind directive would name the opposite role
// from the one `complete` enforces.
func RequiredEvidenceRoles(feature, step string) []orchestration.Role {
	if step == "validate" && !featureUsesAdversarialVerifier(feature) {
		return []orchestration.Role{orchestration.RoleValidationSpec}
	}
	if step == "plan" && !FeatureUsesUnifiedPlanner(feature) {
		return []orchestration.Role{orchestration.RoleBigThinker, orchestration.RoleFeatureSpecial}
	}
	return orchestration.RequiredRolesForFeature(feature, step)
}

// alternateContractRoles returns the roles step requires under the contract pin
// this workflow did NOT take — the other arm of each branch above, kept
// adjacent to it so the two can never drift.
//
// Its only consumer is the handoff-chain check: handing off names a STEP, and
// a workflow whose evidence predates that check carries the name the old
// prefill seeded, which is that step's occupant under the other pin. Nothing
// else may use it — the roles a step actually REQUIRES are always
// RequiredEvidenceRoles.
func alternateContractRoles(feature, step string) []orchestration.Role {
	switch step {
	case "validate":
		if featureUsesAdversarialVerifier(feature) {
			return []orchestration.Role{orchestration.RoleValidationSpec}
		}
	case "plan":
		if FeatureUsesUnifiedPlanner(feature) {
			return []orchestration.Role{orchestration.RoleBigThinker, orchestration.RoleFeatureSpecial}
		}
	default:
		return nil
	}
	return orchestration.RequiredRolesForFeature(feature, step)
}

// annotatePlanContract explains WHY an unpinned workflow's failing plan gate
// names the legacy pair (D2). It wraps ValidateRoles' error rather than forking
// the validator, and deliberately does not offer "planner" as an alternative:
// naming a set this workflow can never satisfy is worse than naming none.
func annotatePlanContract(feature, step string, err error) error {
	if err == nil || step != "plan" || FeatureUsesUnifiedPlanner(feature) {
		return err
	}
	return fmt.Errorf("%w | contract: this workflow predates %q, so its plan step requires the complete legacy big-thinker + feature-specialist pair", err, PlanContractUnified)
}

func strictOrchestrationEnabled(feature string) bool {
	wf, err := Load(feature)
	if err != nil {
		return false
	}
	return wf.OrchestrationMode == StrictOrchestrationMode
}
