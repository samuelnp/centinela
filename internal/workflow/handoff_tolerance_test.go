package workflow

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// hoChain builds tc's repo, writes tc.role evidence carrying got, and runs the
// chain check exactly as validateOrchestration does.
func hoChain(t *testing.T, tc hoCase, got string) error {
	t.Helper()
	hoRepo(t, "f", tc)
	hoEvidence(t, "f", tc.role, got)
	return validateHandoffChain("f", tc.step, RequiredEvidenceRoles("f", tc.step))
}

// hoAssert runs one candidate value and reports whether the verdict matched.
func hoAssert(t *testing.T, tc hoCase, got string, accept bool) {
	t.Helper()
	t.Run(got, func(t *testing.T) {
		err := hoChain(t, tc, got)
		if accept && err != nil {
			t.Fatalf("handoffTo %q must be in-chain: %v", got, err)
		}
		if !accept && err == nil {
			t.Fatalf("handoffTo %q must be refused", got)
		}
	})
}

// outOfChainValues are the values the gate must refuse at a next-step hop into
// validate: every OTHER step's occupant, the terminal literal, the
// out-of-band steward verdicts, case and whitespace variants, and junk.
var outOfChainValues = []string{
	"qa-senior", "documentation-specialist", "planner", "senior-engineer",
	"ux-ui-specialist", "big-thinker", "feature-specialist", "merge-steward",
	"production-readiness", "complete", "user", "banana",
	"GATEKEEPER", "Gatekeeper", " gatekeeper", "gatekeeper ", "gatekeeper\n",
	"  ", strings.Repeat("g", 2000),
}

// TestHandoffToleranceIsStepScoped pins what the alternate-pin tolerance must
// REFUSE. It exists so evidence seeded before this gate — which named the
// right STEP under the other contract pin — keeps completing; it must never
// become a way to name a role from a different step and skip one.
func TestHandoffToleranceIsStepScoped(t *testing.T) {
	tc := hoCase{step: "tests", role: orchestration.RoleQASeniorEngineer}
	for _, in := range []string{"gatekeeper", "validation-specialist"} {
		hoAssert(t, tc, in, true)
	}
	for _, out := range outOfChainValues {
		hoAssert(t, tc, out, false)
	}
}

// TestHandoffToleranceDisabledForSameStepAndTerminal is the stated boundary,
// and the deliberate answer to the old prefill's two retroactive breakages.
//
// The tolerance is NOT widened to cover them. `senior-engineer -> qa-senior`
// on a user-facing feature does not misname the successor step's occupant —
// it names a LATER step's role, i.e. it asserts the same-step ux-ui-specialist
// hop does not exist. Accepting it would hand back exactly the property that
// makes the gate meaningful: that a required same-step role cannot be skipped.
// `gatekeeper -> documentation-specialist` on an internal feature is the same
// shape at the other end — a handoff to a role this workflow requires no
// evidence from at all. Both fail with the executable remedy already printed.
func TestHandoffToleranceDisabledForSameStepAndTerminal(t *testing.T) {
	sameStep := hoCase{userFacing: true, step: "code", role: orchestration.RoleSeniorEngineer}
	hoAssert(t, sameStep, "qa-senior", false)
	hoAssert(t, sameStep, "ux-ui-specialist", true)

	terminal := hoCase{step: "validate", role: orchestration.RoleGatekeeper}
	hoAssert(t, terminal, "documentation-specialist", false)
	hoAssert(t, terminal, TerminalHandoff, true)
}
