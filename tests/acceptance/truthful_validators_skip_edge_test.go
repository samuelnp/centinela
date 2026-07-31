// Acceptance: specs/truthful-validators.feature
//
// Section B (part 2) — the honest edges: a package-level "no test files" skip
// is not a scenario skip, a clean run passes silently, an unrecognized report
// shape warns rather than fails or passes, a summary-shaped substring inside a
// log line is not matched, and a non-zero exit always wins over any skip text.
package acceptance_test

import (
	"strings"
	"testing"
)

// Scenario: A package with no test files is not a skipped scenario
func TestTV_Skip_PackageLevelNoTestFilesIsNotAScenarioSkip(t *testing.T) {
	jsonLines := `{"Action":"run","Package":"p/a","Test":"TestX"}\n` +
		`{"Action":"pass","Package":"p/a","Test":"TestX"}\n` +
		`{"Action":"skip","Package":"p/b"}\n`
	out, code := tvValidateWithCmd(t, `printf '`+jsonLines+`' # go test ./...`, "")
	if code != 0 {
		t.Fatalf("a package-level 'no test files' skip must not fail validate\n%s", out)
	}
}

// Scenario: An acceptance run with everything executed and nothing skipped passes
func TestTV_Skip_AllPassedNoSkipWarning(t *testing.T) {
	out, code := tvValidateWithCmd(t,
		`printf '5 scenarios (5 passed)\n' # tests/acceptance`, "")
	if code != 0 {
		t.Fatalf("an all-passed report must pass validate\n%s", out)
	}
	mustNotContain(t, out, "skipped")
}

// Scenario: An unparseable acceptance report is a warning, never a pass and never a failure
func TestTV_Skip_UnparseableReportIsAWarning(t *testing.T) {
	out, code := tvValidateWithCmd(t,
		`printf 'Ran 12 examples, 0 failures\n' # tests/acceptance`, "")
	if code != 0 {
		t.Fatalf("an undetermined report must not fail validate (exit 0), got %d\n%s", code, out)
	}
	mustContain(t, out, "could not be parsed")
	mustContain(t, out, "-json")
}

// Scenario: A summary-shaped line inside a test's own output is not a skip verdict
func TestTV_Skip_SummaryShapedSubstringInsideLogLineIsIgnored(t *testing.T) {
	out, code := tvValidateWithCmd(t,
		`printf 'log: parser saw the string 2 scenarios (1 skipped) inline\n' # tests/acceptance`, "")
	if code != 0 {
		t.Fatalf("a non-anchored summary-shaped substring must not fail validate\n%s", out)
	}
}

// Scenario: A failing exit code is reported as the failure, not as a skip verdict
func TestTV_Skip_FailingExitCodeWinsOverSkipText(t *testing.T) {
	out, code := tvValidateWithCmd(t,
		`printf '3 scenarios (1 skipped)\n'; exit 1 # tests/acceptance`, "")
	if code == 0 {
		t.Fatalf("a non-zero exit must fail validate\n%s", out)
	}
	if strings.Contains(out, "reported 1 skipped") {
		t.Fatalf("the failure must not be re-labelled as a skip verdict\n%s", out)
	}
}
