package workflow

import "github.com/samuelnp/centinela/internal/orchestration"

// TerminalHandoff is what a role writes when no later step of ITS workflow
// requires any evidence — the end of the chain, not a role name.
const TerminalHandoff = "complete"

// ExpectedHandoff derives the successor role that role must hand off to, from
// the workflow's OWN contract rather than a hardcoded five-role list:
//
//  1. the next required role WITHIN the same step, when one follows it there
//     (senior-engineer -> ux-ui-specialist on a user-facing feature, or the
//     legacy big-thinker -> feature-specialist plan pair);
//  2. otherwise the first required role of the next step that requires any
//     evidence at all;
//  3. otherwise TerminalHandoff.
//
// Deriving it is what keeps the gate from being retroactively brittle: the
// archetypes that omit steps (hotfix, spike), the contract pins that swap a
// step's role (adversarial-v1, planner-v1), and the internal-feature docs step
// that requires no role all fall out of RequiredEvidenceRoles and
// OrderedSteps, which already carry that knowledge.
//
// ok is false when the feature has no workflow state on disk. There is then no
// contract to derive from, and inventing one would be the same fabrication
// this feature exists to stop.
func ExpectedHandoff(feature, step string, role orchestration.Role) (string, bool) {
	wf, err := Load(feature)
	if err != nil {
		return "", false
	}
	want, _ := handoffTarget(wf, feature, step, role)
	return want, true
}

// handoffTarget resolves the successor AND reports whether the hop stays
// inside step. Callers need the distinction: a same-step hop names one
// unambiguous role, while a next-step hop names a STEP whose occupant depends
// on which contract the workflow pinned.
func handoffTarget(wf *Workflow, feature, step string, role orchestration.Role) (want string, sameStep bool) {
	roles := RequiredEvidenceRoles(feature, step)
	for i, r := range roles {
		if r != role {
			continue
		}
		if i+1 < len(roles) {
			return string(roles[i+1]), true
		}
		break
	}
	next, _ := nextChainStep(wf, feature, step)
	return next, false
}

// nextChainStep walks forward from step and returns the first required role of
// the first LATER step that requires any evidence, with that step's name.
//
// Steps requiring none are SKIPPED rather than treated as terminal: an
// archetype that omits a step and an internal feature whose docs step needs no
// role evidence are the same "nothing to hand to here" case, so neither needs
// special-casing. A step outside OrderedSteps (the out-of-band merge step) has
// no successor at all and terminates.
func nextChainStep(wf *Workflow, feature, step string) (role, nextStep string) {
	order := wf.OrderedSteps()
	for i, s := range order {
		if s != step {
			continue
		}
		for _, candidate := range order[i+1:] {
			if roles := RequiredEvidenceRoles(feature, candidate); len(roles) > 0 {
				return string(roles[0]), candidate
			}
		}
		break
	}
	return TerminalHandoff, ""
}
