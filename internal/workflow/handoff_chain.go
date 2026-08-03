package workflow

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// validateHandoffChain checks each required role's handoffTo against the
// successor this workflow's own contract derives, so a value like "banana" or
// "orchestrator" can no longer complete a step that documents a chain.
//
// Unreadable evidence is skipped in SILENCE: orchestration.ValidateRoles has
// already reported it, and a second message for one missing file buries the
// remedy. An empty handoffTo is likewise its rule, not this one.
func validateHandoffChain(feature, step string, roles []orchestration.Role) error {
	for _, role := range roles {
		got, err := orchestration.ReadHandoffTo(orchestration.JSONPath(feature, role))
		if err != nil {
			continue
		}
		if err := CheckHandoffTo(feature, step, role, got); err != nil {
			return err
		}
	}
	return nil
}

// CheckHandoffTo answers for ONE role the question `centinela complete` asks of
// a whole step: is got the successor this workflow's own contract derives?
//
// It is exported so the pre-flight `centinela evidence validate` can ask the
// SAME question. A self-check that reports "evidence ok" for evidence the
// completion gate then refuses is a false success one layer up — precisely the
// signal this feature exists to remove — so the two must not be able to
// disagree.
//
// An empty handoffTo and an absent workflow are both nil here: the first is
// ValidateEvidence's rule, the second means there is no contract to derive
// from, and inventing one would be the fabrication this gate exists to stop.
func CheckHandoffTo(feature, step string, role orchestration.Role, got string) error {
	if got == "" {
		return nil
	}
	wf, err := Load(feature)
	if err != nil {
		return nil
	}
	want, sameStep := handoffTarget(wf, feature, step, role)
	if acceptsHandoff(feature, wf, step, got, want, sameStep) {
		return nil
	}
	return fmt.Errorf("evidence handoffTo for %q is %q, but this workflow's contract makes %q its successor — fix with: centinela evidence set %s %s handoffTo %s",
		role, got, want, feature, role, want)
}

// acceptsHandoff decides whether got is in-chain.
//
// A same-step hop and a terminal handoff each name exactly one legal value. A
// NEXT-step hop is looser by design: handoffTo names the next step's occupant,
// and two steps have two legal occupants depending on which contract the
// workflow pinned (validate: gatekeeper | validation-specialist; plan:
// planner | the legacy big-thinker pair). Evidence written before this gate
// existed carries whichever name the old prefill seeded — the right STEP, so
// in-chain, and retro-failing it would break exactly the in-flight workflows
// the brief says must keep completing. A role belonging to any OTHER step is
// still refused, as is a value that is not a role at all.
func acceptsHandoff(feature string, wf *Workflow, step, got, want string, sameStep bool) bool {
	if got == want {
		return true
	}
	if sameStep || want == TerminalHandoff {
		return false
	}
	_, nextStep := nextChainStep(wf, feature, step)
	for _, alt := range alternateContractRoles(feature, nextStep) {
		if got == string(alt) {
			return true
		}
	}
	return false
}
