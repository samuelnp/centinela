// Acceptance: specs/adversarial-validate-verifier.feature
//
// Strict-orchestration helpers. The other AVV fixtures deliberately run
// non-strict so evidence JSON is never required — but strict is the mode this
// repository actually runs, and it is where the directive and the completion
// gate can disagree about which role's evidence to produce.
package acceptance_test

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// avvSeedStrict writes a workflow at the validate step in strict mode.
func avvSeedStrict(t *testing.T, dir, feature, contract string) {
	t.Helper()
	body := fmt.Sprintf(`{"feature":%q,"currentStep":"validate",`+
		`"stepOrder":["plan","code","tests","validate","docs"],"steps":{},`+
		`"validateContract":%q,"orchestrationMode":"strict-subagents-v1"}`, feature, contract)
	mustWrite(t, filepath.Join(dir, ".workflow", feature+".json"), body)
}

// avvDirective runs `hook orchestration` and returns its output.
func avvDirective(t *testing.T, bin, dir string) string {
	t.Helper()
	cmd := exec.Command(bin, "hook", "orchestration")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader("{}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hook orchestration failed: %v\n%s", err, out)
	}
	return string(out)
}

// avvRequiredEvidence extracts the exact evidence paths the directive tells
// the operator to produce, so a test can assert the GATE accepts that set.
func avvRequiredEvidence(t *testing.T, directive, feature string) []string {
	t.Helper()
	marker := "Required evidence before centinela complete " + feature + ": "
	idx := strings.Index(directive, marker)
	if idx < 0 {
		t.Fatalf("directive names no required evidence for %q: %s", feature, directive)
	}
	line := directive[idx+len(marker):]
	line = strings.SplitN(line, "\n", 2)[0]
	paths := strings.Split(line, ", ")
	for i := range paths {
		paths[i] = strings.TrimSpace(paths[i])
	}
	return paths
}

// avvWriteEvidence materialises the evidence pair the directive asked for,
// deriving each role from its own filename and authoring it through the
// binary's own `evidence` commands — never hand-written JSON, so the fixture
// can never drift from the schema the gate enforces.
//
// Callers MUST run this BEFORE writing the gatekeeper report: for the
// gatekeeper role the companion .md IS the report path, and `evidence init`
// drops a stub over whatever is there.
func avvWriteEvidence(t *testing.T, bin, dir, feature string, paths []string) {
	t.Helper()
	report := ".workflow/" + feature + "-gatekeeper.md"
	for _, rel := range paths {
		if !strings.HasSuffix(rel, ".json") {
			continue
		}
		role := strings.TrimPrefix(strings.TrimSuffix(filepath.Base(rel), ".json"), feature+"-")
		for _, args := range [][]string{
			{"evidence", "init", feature, role},
			{"evidence", "append", feature, role, "inputs", "docs/plans/" + feature + ".md"},
			{"evidence", "append", feature, role, "outputs", report},
		} {
			if out, code := runCent(t, bin, dir, args...); code != 0 {
				t.Fatalf("%v failed (%d): %s", args, code, out)
			}
		}
	}
}

// avvValidateCommands gives the fixture a trivially passing validate command
// so claim verification has ground truth to compare against.
func avvValidateCommands(t *testing.T, dir string) {
	t.Helper()
	mustWrite(t, filepath.Join(dir, "centinela.toml"), "[validate]\ncommands=[\"true\"]\n")
}
