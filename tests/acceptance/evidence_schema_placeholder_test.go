package acceptance_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/evidence-schema-skeleton-legacy-handoff.feature
// Scenario: No-feature path — CWD resolves nothing
// Scenario: No-feature path — ambiguous CWD (parallel sessions) never guesses
// PLUS the case the spec does not state: a `.worktrees/<x>` segment whose
// workflow state is not readable is ALSO the no-feature path. Answering from
// the static legacy chain there prints a stale successor beside a real-looking
// slug — the failure this feature exists to remove, wearing better clothes.
func TestEvidenceSchemaPlaceholderWhenNothingResolves(t *testing.T) {
	cases := []struct {
		name    string
		dir     func(t *testing.T) string
		absent  []string
		wantErr bool
	}{
		{name: "zero active workflows", dir: func(t *testing.T) string { return schemaPlainRepo(t) }},
		{
			name:   "two active never guesses",
			dir:    func(t *testing.T) string { return schemaPlainRepo(t, "demo-a", "demo-b") },
			absent: []string{"demo-a", "demo-b"},
		},
		{name: "worktree segment with no workflow state", dir: schemaStatelessWorktree},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, errOut, code := schemaRun(t, tc.dir(t), "evidence", "schema", "gatekeeper")
			if code != 0 {
				t.Fatalf("exit %d: %s", code, errOut)
			}
			feature, handoff := schemaFields(t, out)
			if feature != schemaSlugSlot || handoff != schemaRoleSlot {
				t.Fatalf("got %q/%q, want both slots unfilled", feature, handoff)
			}
			for _, leak := range tc.absent {
				if strings.Contains(out, leak) {
					t.Fatalf("candidate %q leaked into the output:\n%s", leak, out)
				}
			}
			if !json.Valid([]byte(out)) {
				t.Fatalf("output is not valid JSON:\n%s", out)
			}
		})
	}
}

// Acceptance: Scenario: Unknown role is rejected before any derivation happens
// ("and nothing is printed to stdout") — extended to the arity errors, because
// the same capture path swallows all three straight into a prompt.
func TestEvidenceSchemaErrorPathsWriteNothingToStdout(t *testing.T) {
	dir := schemaPlainRepo(t)
	cases := [][]string{
		{"evidence", "schema", "bogus"},
		{"evidence", "schema"},
		{"evidence", "schema", "gatekeeper", "extra"},
	}
	for _, args := range cases {
		out, _, code := schemaRun(t, dir, args...)
		if code == 0 {
			t.Fatalf("%v must fail", args)
		}
		if len(out) != 0 {
			t.Fatalf("%v wrote %d bytes to stdout: %q", args, len(out), out)
		}
	}
}

// The success path must keep stdout clean of anything but the payload: agents
// pipe it straight into a file. (stderr is asserted separately — a warning
// there is a known, deferred wart of the workflow scan, not this command's.)
func TestEvidenceSchemaSuccessStdoutIsExactlyJSON(t *testing.T) {
	root := schemaRepo(t, "demo-internal", schemaCanonicalOrder, schemaModernPins, "# demo\n", "validate")
	out, _, code := schemaRun(t, filepath.Join(root, "internal", "evidence"), "evidence", "schema", "gatekeeper")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.HasPrefix(out, "{") || !strings.HasSuffix(out, "}\n") || !json.Valid([]byte(out)) {
		t.Fatalf("stdout is not exactly one JSON document:\n%q", out)
	}
}

// schemaPlainRepo returns a directory with no `.worktrees` segment and one
// active workflow file per feature named.
func schemaPlainRepo(t *testing.T, features ...string) string {
	t.Helper()
	dir := t.TempDir()
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	if err := os.MkdirAll(filepath.Join(dir, ".workflow"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range features {
		schemaWrite(t, dir, ".workflow/"+f+".json",
			`{"feature":"`+f+`","currentStep":"code","steps":{},`+
				schemaModernPins+`"stepOrder":[`+schemaCanonicalOrder+`]}`)
	}
	return dir
}

// schemaStatelessWorktree returns a path with a `.worktrees/<x>` segment that
// no workflow backs — a fabricated or stale checkout.
func schemaStatelessWorktree(t *testing.T) string {
	t.Helper()
	base := t.TempDir()
	if r, err := filepath.EvalSymlinks(base); err == nil {
		base = r
	}
	dir := filepath.Join(base, ".worktrees", "not-a-feature", "deep")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}
