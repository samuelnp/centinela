package workflow

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// The same-step guard is only OBSERVABLE where the successor step is one the
// tolerance covers, and the bootstrap order is the single shipped order in
// which the code step's successor is validate. Without the guard a user-facing
// senior-engineer could name a validate-step role and skip the required
// ux-ui-specialist outright — the exact step-skip the tolerance must not enable.
func TestHandoffSameStepGuardBlocksSkippingIntoValidate(t *testing.T) {
	tc := hoCase{userFacing: true, order: BootstrapStepOrder, step: "code",
		role: orchestration.RoleSeniorEngineer}
	for _, got := range []string{"gatekeeper", "validation-specialist", "qa-senior"} {
		hoAssert(t, tc, got, false)
	}
	hoAssert(t, tc, "ux-ui-specialist", true)
}

// The refusal must carry a command an operator can paste, because that command
// IS the migration story for evidence the old prefill seeded.
func TestHandoffErrorNamesExecutableRemedy(t *testing.T) {
	err := hoChain(t, hoCase{userFacing: true, step: "code", role: orchestration.RoleSeniorEngineer}, "qa-senior")
	if err == nil {
		t.Fatal("stale same-step prefill must be refused")
	}
	want := "centinela evidence set f senior-engineer handoffTo ux-ui-specialist"
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error must name %q, got %v", want, err)
	}
}
