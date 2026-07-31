// Acceptance: specs/truthful-validators.feature
//
// Section B (part 1) — acceptance-classified validate commands that report
// skipped/undefined/zero scenarios fail validate. Fixtures use `sh -c` fakes
// standing in for real cucumber/godog/go-test runs; the acceptance-triggering
// substring (e.g. "tests/acceptance") is a trailing shell comment so the
// classifier matches without the comment affecting execution.
package acceptance_test

import "testing"

// tvValidateWithCmd writes a minimal centinela.toml whose one validate command
// is cmd, then runs `centinela validate` in a fresh scratch dir (no git repo,
// so diff-aware degrades to full scan and every built-in gate but G1 stays off).
func tvValidateWithCmd(t *testing.T, cmd, policy string) (string, int) {
	t.Helper()
	bin := buildCent(t)
	dir := t.TempDir()
	toml := "[validate]\ncommands = [" + tvQuote(cmd) + "]\n"
	if policy != "" {
		toml += "acceptance_skip_policy = " + tvQuote(policy) + "\n"
	}
	writeFile(t, dir, "centinela.toml", toml)
	return runCent(t, bin, dir, "validate")
}

func tvQuote(s string) string {
	out := "\""
	for _, r := range s {
		if r == '"' || r == '\\' {
			out += "\\"
		}
		out += string(r)
	}
	return out + "\""
}

// Scenario: A cucumber run reporting a skipped scenario fails validate
func TestTV_Skip_CucumberSkipFailsValidate(t *testing.T) {
	out, code := tvValidateWithCmd(t,
		`printf '3 scenarios (1 skipped, 2 passed)\n' # tests/acceptance`, "")
	if code == 0 {
		t.Fatalf("a reported skip must fail validate\n%s", out)
	}
	mustContain(t, out, "1 skipped")
}

// Scenario: A godog run whose steps are all undefined fails validate
func TestTV_Skip_GodogUndefinedFailsValidate(t *testing.T) {
	out, code := tvValidateWithCmd(t,
		`printf '2 scenarios (2 undefined)\n' # tests/acceptance`, "")
	if code == 0 {
		t.Fatalf("all-undefined must fail validate\n%s", out)
	}
	mustContain(t, out, "2 undefined")
}

// Scenario: A run that executed no scenarios at all fails validate
func TestTV_Skip_ZeroScenariosFailsValidate(t *testing.T) {
	out, code := tvValidateWithCmd(t,
		`printf '0 scenarios\n' # tests/acceptance`, "")
	if code == 0 {
		t.Fatalf("zero scenarios must fail validate\n%s", out)
	}
	mustContain(t, out, "no scenarios")
}

// Scenario: A Go acceptance test that calls t.Skip fails validate
//
// The report body lives in a SCRIPT, not inline: an inline payload naming
// `tests/acceptance` would also put that literal into the command string, which
// is what the classifier reads — the command would scope to the acceptance tier
// and no attribution would run at all. Via a script, the command really is
// whole-repo (`# go test ./...`) and the skip really does have to be
// attributable to the acceptance package for the rule to fire.
// TestTV_Skip_WholeRepoUnitTierSkipStaysGreen pins the other direction.
func TestTV_Skip_GoJSONTestLevelSkipFailsValidate(t *testing.T) {
	jsonLines := `{"Action":"run","Package":"gv/tests/acceptance","Test":"TestX"}\n` +
		`{"Action":"skip","Package":"gv/tests/acceptance","Test":"TestX"}\n`
	out, code := tvValidateWithScript(t, "sh ./run.sh # go test ./...", jsonLines)
	if code == 0 {
		t.Fatalf("a test-level skip in go test -json must fail validate\n%s", out)
	}
	mustContain(t, out, "1 skipped")
}
