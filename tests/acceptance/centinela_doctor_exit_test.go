package acceptance_test

import (
	"strings"
	"testing"
)

// Acceptance: specs/centinela-doctor.feature

// Scenario: Any ERROR check causes exit code 1
func TestDoctorAnyErrorExitsOne(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	writeFile(t, dir, ".workflow/feat-qa-senior.json.tmp", "{}") // evidence ERROR
	_, code := runDoctor(t, dir)
	if code != 1 {
		t.Fatalf("an ERROR check must exit 1, got %d", code)
	}
}

// Scenario: Only WARN checks present causes exit code 0
func TestDoctorOnlyWarnExitsZero(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	writeFile(t, dir, "centinela.toml", "[verify]\nverify_timeout = 60\n") // config WARN
	out, code := runDoctor(t, dir)
	if code != 0 {
		t.Fatalf("WARN-only must exit 0, got %d\n%s", code, out)
	}
	if strings.Contains(out, "✗") {
		t.Fatalf("there must be no ERROR lines:\n%s", out)
	}
}

// Scenario: Summary line always present and reflects actual check counts
func TestDoctorSummaryReflectsCounts(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	out, _ := runDoctor(t, dir)
	last := lastLine(out)
	if !strings.Contains(last, "ok,") || !strings.Contains(last, "warn,") || !strings.Contains(last, "error") {
		t.Fatalf("summary line malformed: %q", last)
	}
}

// Scenario: Output is deterministic — check order is fixed regardless of finding severity
func TestDoctorDeterministicOutput(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	writeFile(t, dir, "centinela.toml", "[verify]\nverify_timeout = 60\n")
	a, _ := runDoctor(t, dir)
	b, _ := runDoctor(t, dir)
	if a != b {
		t.Fatalf("two runs on identical state must be byte-identical:\nA:\n%s\nB:\n%s", a, b)
	}
}

// Scenario: Non-TTY output is plain and parseable with no spinner or ANSI codes
func TestDoctorNonTTYNoANSI(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	out, _ := runDoctor(t, dir)
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("non-TTY output must contain no ANSI escapes:\n%q", out)
	}
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if line == "" {
			continue
		}
		ok := strings.HasPrefix(line, "✓") || strings.HasPrefix(line, "⚠") ||
			strings.HasPrefix(line, "✗") || strings.HasPrefix(line, "  ") ||
			strings.Contains(line, "ok,")
		if !ok {
			t.Fatalf("unexpected line shape: %q", line)
		}
	}
}

func lastLine(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	return lines[len(lines)-1]
}
