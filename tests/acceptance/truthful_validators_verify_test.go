// Acceptance: specs/truthful-validators.feature
//
// Section D — `centinela verify` applies the same acceptance-skip rule as
// `centinela validate`. The workflow fixture is built through the real
// workflow.New/Save API (chdir'd into the scratch dir) so its shape can never
// drift from what `centinela verify` actually loads; the command itself is
// still driven through the compiled binary.
package acceptance_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// tvVerifyFixture builds a scratch repo at the "tests" step with one
// qa-senior evidence file whose only live claim is tests-pass (Coverage nil,
// Outputs and EdgeCases empty, so the other three claim checks Skip and
// cannot mask the assertion under test), plus a single validate command.
func tvVerifyFixture(t *testing.T, cmd string) (bin, dir, feature string) {
	t.Helper()
	bin = buildCent(t)
	dir = t.TempDir()
	feature = "alpha"
	t.Chdir(dir)
	if err := os.MkdirAll(filepath.Join(dir, workflow.WorkflowDir), 0o755); err != nil {
		t.Fatal(err)
	}
	wf := workflow.New(feature)
	wf.CurrentStep = "tests"
	if err := workflow.Save(wf); err != nil {
		t.Fatal(err)
	}
	writeFile(t, dir, "centinela.toml", "[validate]\ncommands = ["+tvQuote(cmd)+"]\n")
	evidence := `{"feature":"` + feature + `","step":"tests","role":"qa-senior",` +
		`"status":"done","generatedAt":"2026-01-01T00:00:00Z",` +
		`"inputs":["i"],"outputs":[],"edgeCases":[],"handoffTo":"validation-specialist"}`
	writeFile(t, dir, ".workflow/"+feature+"-qa-senior.json", evidence)
	return bin, dir, feature
}

// Scenario: The claim verifier rejects a tests-pass claim whose acceptance run skipped
func TestTV_Verify_RejectsTestsPassClaimOnAcceptanceSkip(t *testing.T) {
	bin, dir, feature := tvVerifyFixture(t,
		`printf '3 scenarios (1 skipped, 2 passed)\n' # tests/acceptance`)
	out, code := runCent(t, bin, dir, "verify", feature)
	if code == 0 {
		t.Fatalf("a claimed tests-pass whose acceptance run skipped must fail verification\n%s", out)
	}
	mustContain(t, out, "skipped")
}

// Scenario: The claim verifier does not invent a failure it cannot prove
func TestTV_Verify_UnparseableStaysPass(t *testing.T) {
	bin, dir, feature := tvVerifyFixture(t,
		`printf 'Ran 12 examples, 0 failures\n' # tests/acceptance`)
	out, code := runCent(t, bin, dir, "verify", feature)
	if code != 0 {
		t.Fatalf("an unprovable skip must not fail verification\n%s", out)
	}
	mustContain(t, out, "could not be parsed")
}
