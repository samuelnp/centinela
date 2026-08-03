package acceptance_test

import (
	"strings"
	"testing"
)

// Acceptance: specs/evidence-schema-skeleton-legacy-handoff.feature
// Scenario: No-feature path — "pasting this JSON verbatim ... reports a
// handoffTo issue naming the true successor and a fix command".
//
// That scenario is TRUE ONLY for a role the workflow's contract requires: the
// chain check iterates RequiredEvidenceRoles, so a non-required role's
// handoffTo is never inspected and the literal placeholder passes. Both halves
// are pinned here — the loud one because it is the guarantee, the quiet one
// because the spec and the help text used to claim it could not happen. The
// gate itself is deliberately NOT widened (out of scope; deferred as
// handoff-gate-skips-nonrequired-role-evidence).
func TestEvidenceSchemaPlaceholderPastedVerbatim(t *testing.T) {
	cases := []struct {
		name, pins string
		wantLoud   bool
	}{
		{"gatekeeper is required under the adversarial pin", schemaModernPins, true},
		{"gatekeeper is not required on a legacy-pinned workflow", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := schemaRepo(t, "demo", schemaCanonicalOrder, tc.pins, "# demo\n", "validate")
			schemaSeedGatekeeperEvidence(t, root, schemaRoleSlot)
			out, errOut, code := schemaRun(t, root, "evidence", "validate", "demo")
			if !tc.wantLoud {
				if code != 0 || !strings.Contains(out, "evidence ok") {
					t.Fatalf("expected the placeholder to be accepted, got exit %d\n%s%s", code, out, errOut)
				}
				return
			}
			if code == 0 {
				t.Fatalf("required role must refuse the placeholder, got exit 0:\n%s", out)
			}
			for _, want := range []string{"handoffTo", schemaRoleSlot, "complete", "centinela evidence set demo gatekeeper handoffTo"} {
				if !strings.Contains(errOut, want) {
					t.Fatalf("refusal must contain %q, got:\n%s", want, errOut)
				}
			}
		})
	}
}

// schemaSeedGatekeeperEvidence writes a gatekeeper evidence file that is
// complete in every field EXCEPT handoffTo, which is set to handoff — so the
// only thing `evidence validate` can object to is the handoff value.
func schemaSeedGatekeeperEvidence(t *testing.T, root, handoff string) {
	t.Helper()
	schemaWrite(t, root, "docs/plans/demo.md", "# plan\n")
	schemaWrite(t, root, ".workflow/demo-gatekeeper.md", "**Status:** SAFE\n")
	steps := [][]string{
		{"evidence", "init", "demo", "gatekeeper"},
		{"evidence", "append", "demo", "gatekeeper", "inputs", "docs/plans/demo.md"},
		{"evidence", "append", "demo", "gatekeeper", "outputs", ".workflow/demo-gatekeeper.md"},
		{"evidence", "append", "demo", "gatekeeper", "edgeCases", "placeholder pasted verbatim"},
		{"evidence", "set", "demo", "gatekeeper", "handoffTo", handoff},
	}
	for _, args := range steps {
		if out, errOut, code := schemaRun(t, root, args...); code != 0 {
			t.Fatalf("%v exit %d:\n%s%s", args, code, out, errOut)
		}
	}
}
