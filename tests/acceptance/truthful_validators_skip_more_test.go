// Acceptance: specs/truthful-validators.feature
//
// Section B (part 3) — a truncated report never derives a skip verdict, a
// non-acceptance command is never touched by skip analysis, an empty command
// list runs nothing, and the moved classifier keeps its old verdicts.
package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Scenario: A truncated acceptance report is undetermined, not a skip verdict
//
// `centinela validate` has no wall-clock timeout of its own (that concept
// lives in internal/verify's Deps.timeout for `centinela verify`); the
// closest honest analogue here is a command that is killed mid-write, leaving
// a partial summary line and a non-zero exit. Exit-code-wins means the
// partial text can never be read as a skip verdict either way.
func TestTV_Skip_TruncatedReportNeverDerivesASkipVerdict(t *testing.T) {
	out, code := tvValidateWithCmd(t,
		`printf '3 scenarios (1 sk'; exit 1 # tests/acceptance`, "")
	if code == 0 {
		t.Fatalf("the killed/failing command must still fail validate\n%s", out)
	}
	if strings.Contains(out, "skipped, ") {
		t.Fatalf("a partial summary must never be read as a skip verdict\n%s", out)
	}
}

// Scenario: A skipping unit or integration command is never failed by skip detection
func TestTV_Skip_NonAcceptanceCommandNeverFailedBySkipDetection(t *testing.T) {
	out, code := tvValidateWithCmd(t, `printf 'unit tier: SKIP TestFoo (0.00s)\n'`, "")
	if code != 0 {
		t.Fatalf("a non-acceptance command reporting skips must still pass\n%s", out)
	}
}

// Scenario: A project with no configured validate commands is unaffected
func TestTV_Skip_NoConfiguredCommandsIsUnaffected(t *testing.T) {
	bin := buildCent(t)
	dir := t.TempDir()
	writeFile(t, dir, "centinela.toml", "[validate]\ncommands = []\n")
	out, code := runCent(t, bin, dir, "validate")
	if code != 0 {
		t.Fatalf("an empty command list must not fail validate\n%s", out)
	}
	if strings.Contains(out, "Validate Commands") {
		t.Fatalf("no commands were configured, so no command section should render\n%s", out)
	}
}

// Scenario: The acceptance classifier is unchanged by this feature
//
// Structural pin: internal/workflow's predicate must delegate to the leaf
// byte-for-byte rather than reimplement it — the proof this feature MOVED the
// classification instead of broadening it.
func TestTV_Skip_ClassifierDelegationIsUnchanged(t *testing.T) {
	path := filepath.Join(repoRoot(t), "internal", "workflow", "validate_tests_acceptance_commands.go")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	src := string(data)
	if !strings.Contains(src, "acceptance.AnyExecutionCommand") {
		t.Fatalf("hasAcceptanceExecutionCommand must delegate to acceptance.AnyExecutionCommand:\n%s", src)
	}
}
