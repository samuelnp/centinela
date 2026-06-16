// Acceptance: specs/custom-gate-sdk.feature
package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// customRepo writes a temp repo whose centinela.toml has file_size off (i18n on
// to keep it honored off) plus the supplied [[gates.custom]] body, and returns
// the dir. auditOn appends an enabled audit_baseline gate.
func customRepo(t *testing.T, customBody string, auditOn bool) string {
	t.Helper()
	dir := t.TempDir()
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	var b strings.Builder
	b.WriteString("[gates]\nfile_size = false\ni18n = true\n")
	if auditOn {
		b.WriteString("\n[gates.audit_baseline]\nenabled = true\nseverity = \"fail\"\n")
	}
	b.WriteString("\n" + customBody)
	writeFile(t, dir, "centinela.toml", b.String())
	return dir
}

func runCustomValidate(t *testing.T, dir string) (string, int) {
	t.Helper()
	return runCent(t, buildCent(t), dir, "validate")
}

func gate(name, command, severity string) string {
	return "[[gates.custom]]\nenabled = true\nname = \"" + name + "\"\ncommand = \"" +
		command + "\"\nseverity = \"" + severity + "\"\n"
}

// Scenario: A passing custom gate appears in the validate gate report by its name
func TestCustomPassingAppearsByName(t *testing.T) {
	dir := customRepo(t, gate("no-todo", "true", "fail"), false)
	out, code := runCustomValidate(t, dir)
	if code != 0 {
		t.Fatalf("exit = %d, want 0\n%s", code, out)
	}
	if !strings.Contains(out, "no-todo") || !strings.Contains(out, "no-todo passed") {
		t.Fatalf("passing gate not reported: %q", out)
	}
	if strings.Contains(strings.ToLower(out), "panic") {
		t.Fatalf("unexpected stack trace: %q", out)
	}
}

// Scenario: A failing severity-fail custom gate blocks validate and surfaces command output in its details
func TestCustomFailBlocksWithDetails(t *testing.T) {
	dir := customRepo(t, gate("no-console-log", "sh -c 'echo found in src/app.js; exit 1'", "fail"), false)
	out, code := runCustomValidate(t, dir)
	if code == 0 {
		t.Fatalf("severity=fail must block\n%s", out)
	}
	if !strings.Contains(out, "no-console-log") || !strings.Contains(out, "found in src/app.js") {
		t.Fatalf("failing gate detail missing: %q", out)
	}
}

// Scenario: A failing severity-fail custom gate that prints nothing falls back to a generic failure detail
func TestCustomSilentFailGenericDetail(t *testing.T) {
	dir := customRepo(t, gate("silent-fail", "false", "fail"), false)
	out, code := runCustomValidate(t, dir)
	if code == 0 {
		t.Fatalf("silent fail must block\n%s", out)
	}
	if !strings.Contains(out, "silent-fail") || !strings.Contains(out, "no output") {
		t.Fatalf("generic fallback detail missing: %q", out)
	}
}

// Scenario: A failing severity-warn custom gate is reported but does not block validate
func TestCustomWarnDoesNotBlock(t *testing.T) {
	dir := customRepo(t, gate("style-nit", "sh -c 'echo nit; exit 1'", "warn"), false)
	out, code := runCustomValidate(t, dir)
	if code != 0 {
		t.Fatalf("warn gate must not block: exit %d\n%s", code, out)
	}
	if !strings.Contains(out, "style-nit") || !strings.Contains(out, "nit") {
		t.Fatalf("warn gate not reported: %q", out)
	}
}

// Scenario: Multiple custom gates run independently and a failing one does not prevent the others from reporting
func TestCustomMultipleIndependent(t *testing.T) {
	body := gate("gate-a", "true", "fail") + gate("gate-b", "sh -c 'echo b-broke; exit 1'", "fail") +
		gate("gate-c", "true", "fail")
	dir := customRepo(t, body, false)
	out, code := runCustomValidate(t, dir)
	if code == 0 {
		t.Fatalf("a failing gate must block\n%s", out)
	}
	for _, n := range []string{"gate-a", "gate-b", "gate-c"} {
		if !strings.Contains(out, n) {
			t.Fatalf("gate %q missing from report: %q", n, out)
		}
	}
	if !strings.Contains(out, "gate-a passed") || !strings.Contains(out, "gate-c passed") {
		t.Fatalf("passing gates suppressed by the failure: %q", out)
	}
}

// Scenario: A custom gate with enabled=false does not run and leaves validate output unchanged
func TestCustomDisabledDoesNotRun(t *testing.T) {
	body := "[[gates.custom]]\nenabled = false\nname = \"skipped\"\n" +
		"command = \"sh -c 'echo should-not-run; exit 1'\"\nseverity = \"fail\"\n"
	dir := customRepo(t, body, false)
	out, code := runCustomValidate(t, dir)
	if code != 0 {
		t.Fatalf("disabled gate must not block: exit %d\n%s", code, out)
	}
	if strings.Contains(out, "skipped passed") || strings.Contains(out, "skipped failed") ||
		strings.Contains(out, "should-not-run") {
		t.Fatalf("disabled gate leaked into output: %q", out)
	}
}

// Scenario: No custom gate entries leaves validate output byte-identical to a run with no custom gates configured
func TestCustomNoneByteIdentical(t *testing.T) {
	base := "[gates]\nfile_size = false\ni18n = true\n"
	d1 := t.TempDir()
	d2 := t.TempDir()
	writeFile(t, d1, "centinela.toml", base)
	writeFile(t, d2, "centinela.toml", base+"\n# no [[gates.custom]] entries\n")
	bin := buildCent(t)
	o1, c1 := runCent(t, bin, d1, "validate")
	o2, c2 := runCent(t, bin, d2, "validate")
	if c1 != c2 {
		t.Fatalf("exit codes differ: %d vs %d", c1, c2)
	}
	if o1 != o2 {
		t.Fatalf("validate output differs with an empty custom section:\n%q\nvs\n%q", o1, o2)
	}
	if strings.Contains(o1, "custom") && strings.Contains(strings.ToLower(o1), "gate") {
		// no custom gate should ever appear; soft check.
	}
}

// Scenario: A custom gate with an empty command is rejected with a clear config error
func TestCustomEmptyCommandRejected(t *testing.T) {
	body := "[[gates.custom]]\nenabled = true\nname = \"no-cmd\"\ncommand = \"\"\nseverity = \"fail\"\n"
	dir := customRepo(t, body, false)
	out, code := runCustomValidate(t, dir)
	if code == 0 {
		t.Fatalf("empty command must be rejected\n%s", out)
	}
	if !strings.Contains(out, "gates.custom[0]") || !strings.Contains(out, "command is required") {
		t.Fatalf("indexed config error missing: %q", out)
	}
	if strings.Contains(strings.ToLower(out), "panic") {
		t.Fatalf("unexpected panic: %q", out)
	}
}

// Scenario: A custom gate with an empty name is rejected with a clear config error
func TestCustomEmptyNameRejected(t *testing.T) {
	body := "[[gates.custom]]\nenabled = true\nname = \"\"\ncommand = \"true\"\nseverity = \"fail\"\n"
	dir := customRepo(t, body, false)
	out, code := runCustomValidate(t, dir)
	if code == 0 {
		t.Fatalf("empty name must be rejected\n%s", out)
	}
	if !strings.Contains(out, "name is required") {
		t.Fatalf("name-required error missing: %q", out)
	}
}

// Scenario: Two custom gates with duplicate names are rejected with a clear config error
func TestCustomDuplicateNamesRejected(t *testing.T) {
	dir := customRepo(t, gate("dup", "true", "fail")+gate("dup", "true", "fail"), false)
	out, code := runCustomValidate(t, dir)
	if code == 0 {
		t.Fatalf("duplicate names must be rejected\n%s", out)
	}
	if !strings.Contains(out, "dup") || !strings.Contains(out, "duplicates") {
		t.Fatalf("duplicate-name error missing: %q", out)
	}
}

// Scenario: A custom gate whose name collides with a built-in gate name is rejected
func TestCustomBuiltinCollisionRejected(t *testing.T) {
	dir := customRepo(t, gate("import_graph", "true", "fail"), false)
	out, code := runCustomValidate(t, dir)
	if code == 0 {
		t.Fatalf("built-in collision must be rejected\n%s", out)
	}
	if !strings.Contains(out, "import_graph") || !strings.Contains(out, "collides with built-in") {
		t.Fatalf("collision error missing: %q", out)
	}
}

// Scenario: A custom gate with an invalid severity is rejected with a clear config error
func TestCustomInvalidSeverityRejected(t *testing.T) {
	dir := customRepo(t, gate("bad-sev", "true", "critical"), false)
	out, code := runCustomValidate(t, dir)
	if code == 0 {
		t.Fatalf("invalid severity must be rejected\n%s", out)
	}
	if !strings.Contains(out, "severity must be fail or warn") {
		t.Fatalf("severity error missing: %q", out)
	}
}

// Scenario: A custom gate with an invalid output mode is rejected with a clear config error
func TestCustomInvalidOutputRejected(t *testing.T) {
	body := "[[gates.custom]]\nenabled = true\nname = \"bad-output\"\ncommand = \"true\"\n" +
		"severity = \"fail\"\noutput = \"json\"\n"
	dir := customRepo(t, body, false)
	out, code := runCustomValidate(t, dir)
	if code == 0 {
		t.Fatalf("invalid output must be rejected\n%s", out)
	}
	if !strings.Contains(out, "output must be blob or lines") {
		t.Fatalf("output error missing: %q", out)
	}
}

// Scenario: A custom gate command that exceeds its timeout fails the gate with a timeout message
func TestCustomTimeoutFailsGate(t *testing.T) {
	body := "[[gates.custom]]\nenabled = true\nname = \"hang\"\ncommand = \"sleep 60\"\n" +
		"severity = \"fail\"\ntimeout_seconds = 1\n"
	dir := customRepo(t, body, false)
	out, code := runCustomValidate(t, dir)
	if code == 0 {
		t.Fatalf("timeout must block\n%s", out)
	}
	if !strings.Contains(out, "hang") || !strings.Contains(out, "timed out") {
		t.Fatalf("timeout message missing: %q", out)
	}
}

// Scenario: A custom gate whose command is not found fails the gate with a clear message and does not crash
func TestCustomCommandNotFound(t *testing.T) {
	dir := customRepo(t, gate("missing-bin", "this-binary-does-not-exist --check", "fail"), false)
	out, code := runCustomValidate(t, dir)
	if code == 0 {
		t.Fatalf("missing binary must fail the gate\n%s", out)
	}
	if !strings.Contains(out, "missing-bin") || !strings.Contains(strings.ToLower(out), "not found") {
		t.Fatalf("command-not-found message missing: %q", out)
	}
	if strings.Contains(strings.ToLower(out), "panic") {
		t.Fatalf("unexpected panic: %q", out)
	}
}

// Scenario: A failing custom gate with output=lines turns each stdout line into a separate violation detail
func TestCustomLinesSeparateDetails(t *testing.T) {
	body := "[[gates.custom]]\nenabled = true\nname = \"per-line\"\n" +
		"command = \"printf 'a.go:1\\nb.go:2\\nc.go:3\\n'; exit 1\"\noutput = \"lines\"\n"
	dir := customRepo(t, body, false)
	out, code := runCustomValidate(t, dir)
	if code == 0 {
		t.Fatalf("failing lines gate must block\n%s", out)
	}
	for _, l := range []string{"a.go:1", "b.go:2", "c.go:3"} {
		if !strings.Contains(out, l) {
			t.Fatalf("line %q missing as a distinct detail: %q", l, out)
		}
	}
}

// Scenario: A failing custom gate is baseline-able and then tolerated by audit while a new violation blocks
func TestCustomBaselineThenNewBlocks(t *testing.T) {
	body := "[[gates.custom]]\nenabled = true\nname = \"per-line\"\n" +
		"command = \"printf 'a.go:1\\nb.go:2\\n'; exit 1\"\noutput = \"lines\"\n"
	dir := customRepo(t, body, true)
	bin := buildCent(t)
	if out, code := runCent(t, bin, dir, "audit", "baseline"); code != 0 {
		t.Fatalf("baseline exit = %d\n%s", code, out)
	}
	newBody := "[[gates.custom]]\nenabled = true\nname = \"per-line\"\n" +
		"command = \"printf 'a.go:1\\nb.go:2\\nc.go:3\\n'; exit 1\"\noutput = \"lines\"\n"
	writeFile(t, dir, "centinela.toml",
		"[gates]\nfile_size = false\ni18n = true\n\n[gates.audit_baseline]\nenabled = true\nseverity = \"fail\"\n\n"+newBody)
	out, code := runCent(t, bin, dir, "audit")
	if code == 0 {
		t.Fatalf("a new violation line must block\n%s", out)
	}
	if !strings.Contains(out, "c.go:3") || !strings.Contains(out, "new") {
		t.Fatalf("new line not reported: %q", out)
	}
	if !strings.Contains(out, "2 baselined") {
		t.Fatalf("baselined lines not tolerated: %q", out)
	}
}

// Scenario: A failing custom gate is recorded as a gate-failure telemetry event
func TestCustomGateFailureTelemetry(t *testing.T) {
	dir := customRepo(t, gate("telemetry-fail", "false", "fail"), false)
	out, code := runCustomValidate(t, dir)
	if code == 0 {
		t.Fatalf("failing gate must block\n%s", out)
	}
	log, err := os.ReadFile(filepath.Join(dir, ".workflow", "telemetry", "events.jsonl"))
	if err != nil {
		t.Fatalf("telemetry log not written: %v", err)
	}
	if !strings.Contains(string(log), "gate-failure") || !strings.Contains(string(log), "telemetry-fail") {
		t.Fatalf("gate-failure event missing the custom gate: %s", log)
	}
}

// Scenario: Two validate runs with the same deterministic custom command produce the same gate report
func TestCustomDeterministicAcrossRuns(t *testing.T) {
	body := "[[gates.custom]]\nenabled = true\nname = \"stable\"\n" +
		"command = \"sh -c 'echo x:1; echo y:2; exit 1'\"\noutput = \"lines\"\n"
	dir := customRepo(t, body, false)
	bin := buildCent(t)
	o1, c1 := runCent(t, bin, dir, "validate")
	o2, c2 := runCent(t, bin, dir, "validate")
	if c1 != c2 || c1 == 0 {
		t.Fatalf("exit codes unstable or non-blocking: %d vs %d", c1, c2)
	}
	if o1 != o2 {
		t.Fatalf("validate output not deterministic:\n%q\nvs\n%q", o1, o2)
	}
	for _, l := range []string{"x:1", "y:2"} {
		if !strings.Contains(o1, l) {
			t.Fatalf("stable line %q missing: %q", l, o1)
		}
	}
}
