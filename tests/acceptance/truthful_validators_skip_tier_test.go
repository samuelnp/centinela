// Acceptance: specs/truthful-validators.feature
//
// Section B (part 3) — the two end-to-end holes the adversarial verifier found:
// a whole-repo command must not fail on ANOTHER tier's legitimate skip, and a
// clean Gherkin summary must not hide a go-level skip printed beside it.
package acceptance_test

import "testing"

// tvValidateWithScript runs `centinela validate` over a one-command config that
// invokes a SCRIPT. The report body therefore lives in the script, not in the
// command string — essential here, because the classifier reads the command
// string and an inline `printf 'tests/acceptance…'` would scope the command to
// the acceptance tier and defeat the whole-repo case under test.
func tvValidateWithScript(t *testing.T, cmd, body string) (string, int) {
	t.Helper()
	bin := buildCent(t)
	dir := t.TempDir()
	writeFile(t, dir, "run.sh", "#!/bin/sh\nprintf '%b' '"+body+"'\n")
	writeFile(t, dir, "centinela.toml", "[validate]\ncommands = ["+tvQuote(cmd)+"]\n")
	return runCent(t, bin, dir, "validate")
}

const tvUnitBlock = `--- SKIP: TestUnitSkip (0.00s)\nok  \tgv/unitpkg\t0.10s\n`
const tvAcceptBlock = `--- SKIP: TestAcceptSkip (0.00s)\nok  \tgv/tests/acceptance\t0.10s\n`

// Scenario: A skipping unit or integration command is never failed by skip detection
//
// `go test -v ./...` executes the acceptance tier, so it is acceptance-
// classified — but it also reports the unit tier's build-tag, -short and
// platform skips. Those must never fail the run, or following this tool's own
// "add -v to make skips detectable" advice would break every project.
func TestTV_Skip_WholeRepoUnitTierSkipStaysGreen(t *testing.T) {
	out, code := tvValidateWithScript(t, "sh ./run.sh # go test -v ./...", tvUnitBlock)
	if code != 0 {
		t.Fatalf("a unit-tier skip must not fail a whole-repo command\n%s", out)
	}
}

// The mirror: the same whole-repo command, the same shape, the skip in the
// ACCEPTANCE tier — still a failure. The fix narrows attribution, it does not
// disarm detection.
func TestTV_Skip_WholeRepoAcceptanceTierSkipStillFails(t *testing.T) {
	out, code := tvValidateWithScript(t, "sh ./run.sh # go test -v ./...", tvAcceptBlock)
	if code == 0 {
		t.Fatalf("an acceptance-tier skip must still fail validate\n%s", out)
	}
	mustContain(t, out, "1 skipped")
	mustContain(t, out, "attributed to tests/acceptance")
}

// Scenario: A Go acceptance test that calls t.Skip fails validate
//
// A run emitting a clean "N scenarios (N passed)" summary AND a go-level
// --- SKIP: is the natural shape of one acceptance package holding both godog
// scenarios and plain Go tests. Taking only the first matching parser rendered
// a silent green — the exact false-green class this feature exists to kill.
func TestTV_Skip_CleanSummaryDoesNotHideAGoSkip(t *testing.T) {
	mixed := `3 scenarios (3 passed)\n--- SKIP: TestGoLevelHidden (0.00s)\n` +
		`ok  \tx/tests/acceptance\t0.42s\n`
	out, code := tvValidateWithScript(t, "sh ./run.sh # tests/acceptance", mixed)
	if code == 0 {
		t.Fatalf("a go skip beside a clean summary must fail validate\n%s", out)
	}
	mustContain(t, out, "1 skipped")
}

// Scenario: A run that executed no scenarios at all fails validate
//
// The union of shapes must not let an unrelated passing signal rescue a suite
// that ran nothing: godog driven from a Go wrapper test prints BOTH a
// `0 scenarios` summary and the wrapper's own `--- PASS:`.
func TestTV_Skip_ZeroScenariosBesideAPassingGoTestStillFails(t *testing.T) {
	wrapped := `0 scenarios\n0 steps\n--- PASS: TestFeatures (0.00s)\n` +
		`PASS\nok  \tgv/tests/acceptance\t0.10s\n`
	out, code := tvValidateWithScript(t, "sh ./run.sh # tests/acceptance", wrapped)
	if code == 0 {
		t.Fatalf("a 0-scenario run must fail even beside a passing Go test\n%s", out)
	}
	mustContain(t, out, "no scenarios")
}

// Scenario: A skipping unit or integration command is never failed by skip detection
//
// The Gherkin half needs the same tier attribution the Go half got: a UNIT
// package printing a scenario summary must not fail the acceptance gate.
func TestTV_Skip_WholeRepoGherkinInAnotherTierStaysGreen(t *testing.T) {
	body := `2 scenarios (1 skipped, 1 passed)\nPASS\nok  \tgv/unitpkg\t0.10s\n` +
		`--- PASS: TestAcceptOK (0.00s)\nPASS\nok  \tgv/tests/acceptance\t0.10s\n`
	out, code := tvValidateWithScript(t, "sh ./run.sh # go test -v ./...", body)
	if code != 0 {
		t.Fatalf("another tier's Gherkin skip must not fail the gate\n%s", out)
	}
}
