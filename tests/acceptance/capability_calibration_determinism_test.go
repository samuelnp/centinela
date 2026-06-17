package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/capability-calibration.feature

// mixedLog is a fixed multi-model log reused by the determinism/sort scenarios.
func mixedLog() []string {
	return calConcat(
		calRepeat(3, func() string { return adv("claude-sonnet-4-6") }), []string{gf("claude-sonnet-4-6")},
		calRepeat(3, func() string { return adv("claude-haiku-4-5") }), []string{gf("claude-haiku-4-5")},
		calRepeat(3, func() string { return adv("") }), []string{gf("")})
}

// Scenario: Two runs on the same log produce byte-identical output
func TestCalDeterministicHuman(t *testing.T) {
	dir := calRepo(t, mixedLog())
	a, _ := runCal(t, dir)
	b, _ := runCal(t, dir)
	if a != b {
		t.Fatalf("human output not byte-identical:\n%s\n---\n%s", a, b)
	}
}

// Scenario: Two --json runs on the same log produce byte-identical JSON output
func TestCalDeterministicJSON(t *testing.T) {
	dir := calRepo(t, mixedLog())
	a, _ := runCal(t, dir, "--json")
	b, _ := runCal(t, dir, "--json")
	if a != b {
		t.Fatalf("json output not byte-identical:\n%s\n---\n%s", a, b)
	}
}

// Scenario: Models are sorted by id ascending with unattributed forced last
func TestCalSortedUnattributedLast(t *testing.T) {
	out, code := runCal(t, calRepo(t, mixedLog()))
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	idxBefore(t, out, "claude-haiku-4-5", "claude-sonnet-4-6")
	idxBefore(t, out, "claude-sonnet-4-6", "unattributed")
}

// Scenario: Non-TTY piped output contains no ANSI escape sequences
func TestCalNonTTYNoANSI(t *testing.T) {
	dir := calRepo(t, calConcat(calRepeat(3, func() string { return adv("claude-sonnet-4-6") }),
		[]string{gf("claude-sonnet-4-6")}))
	out, code := runCal(t, dir)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	// Persist to a file as the scenario describes, then re-read and assert plain.
	f := filepath.Join(dir, "out.txt")
	if err := os.WriteFile(f, []byte(out), 0o644); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(f)
	if strings.Contains(string(b), "\x1b[") {
		t.Fatalf("piped output contains ANSI: %q", b)
	}
}
