package acceptance_test

import (
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/evidence-schema-skeleton-legacy-handoff.feature
// Scenario: Derive with feature — canonical internal feature, worktree CWD
// Scenario: Derive with feature — hotfix archetype has no docs step
// Scenario: Derive with feature — user-facing feature, same-step ux-ui hop
// PLUS the scenario the spec is missing (see the qa-senior report): the same
// three answers must hold when the command is invoked from a SUBDIRECTORY of
// the worktree. None of the 13 prompt lines that call this command tell the
// agent to cd to the root first, so depth > 0 is the normal case, and before
// the fix it printed the stale legacy successor beside a correct feature slug.
func TestEvidenceSchemaDerivesFromAnyDepth(t *testing.T) {
	cases := []struct{ name, feature, order, pins, brief, role, step, want string }{
		{"canonical internal", "demo-internal", schemaCanonicalOrder, schemaModernPins, "# demo\n", "gatekeeper", "validate", "complete"},
		{"hotfix has no docs step", "demo-hotfix", `"code","tests","validate"`, schemaModernPins, "# demo\n", "gatekeeper", "validate", "complete"},
		{"user-facing ux-ui hop", "demo-uxfacing", schemaCanonicalOrder, schemaModernPins, "surface: user-facing\n", "senior-engineer", "code", "ux-ui-specialist"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := schemaRepo(t, tc.feature, tc.order, tc.pins, tc.brief, tc.step)
			for _, dir := range []string{root, filepath.Join(root, "internal", "evidence")} {
				out, errOut, code := schemaRun(t, dir, "evidence", "schema", tc.role)
				if code != 0 {
					t.Fatalf("exit %d from %s: %s", code, dir, errOut)
				}
				feature, handoff := schemaFields(t, out)
				if feature != tc.feature || handoff != tc.want {
					t.Fatalf("from %s: %q/%q, want %q/%q", dir, feature, handoff, tc.feature, tc.want)
				}
				if handoff == "documentation-specialist" && tc.want != "documentation-specialist" {
					t.Fatalf("legacy chain value resurfaced from %s", dir)
				}
			}
		})
	}
}

// Acceptance: the printed handoffTo must equal what `evidence init` prefills —
// one derivation, three callers (schema, init, the completion gate).
func TestEvidenceSchemaAgreesWithEvidenceInit(t *testing.T) {
	root := schemaRepo(t, "demo-internal", schemaCanonicalOrder, schemaModernPins, "# demo\n", "validate")
	sub := filepath.Join(root, "internal", "evidence")
	printed, _, code := schemaRun(t, sub, "evidence", "schema", "gatekeeper")
	if code != 0 {
		t.Fatalf("schema exit %d", code)
	}
	_, handoff := schemaFields(t, printed)
	if _, errOut, code := schemaRun(t, root, "evidence", "init", "demo-internal", "gatekeeper"); code != 0 {
		t.Fatalf("init exit %d: %s", code, errOut)
	}
	written, _, code := schemaRun(t, root, "evidence", "read", "demo-internal", "gatekeeper")
	if code != 0 {
		t.Fatalf("evidence read exit %d", code)
	}
	if !strings.Contains(written, `"handoffTo": "`+handoff+`"`) {
		t.Fatalf("init prefilled something else than %q:\n%s", handoff, written)
	}
}

// Acceptance: Scenario: Merge-steward is out-of-band — never guessed, never
// placeholder'd; and unaffected by derivation when a feature IS resolvable.
func TestEvidenceSchemaMergeStewardAlwaysComplete(t *testing.T) {
	root := schemaRepo(t, "demo-internal", schemaCanonicalOrder, schemaModernPins, "# demo\n", "validate")
	for _, dir := range []string{root, filepath.Join(root, "internal", "evidence"), t.TempDir()} {
		out, _, code := schemaRun(t, dir, "evidence", "schema", "merge-steward")
		if code != 0 {
			t.Fatalf("exit %d from %s", code, dir)
		}
		if _, handoff := schemaFields(t, out); handoff != "complete" {
			t.Fatalf("merge-steward from %s handed off to %q", dir, handoff)
		}
	}
}
