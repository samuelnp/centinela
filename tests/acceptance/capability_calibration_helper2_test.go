package acceptance_test

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

// Acceptance: specs/capability-calibration.feature (helper continuation).

// calSeq is a monotonic counter giving each generated event a distinct,
// ascending RFC3339 timestamp so span ordering is deterministic.
var calSeq int

// calLine builds one JSONL event line of the given type for model (empty model →
// no "model" key, exercising the unattributed bucket and back-compat read).
func calLine(typ, model string) string {
	calSeq++
	ts := fmt.Sprintf("2026-01-%02dT00:00:00Z", (calSeq%27)+1)
	if model == "" {
		return fmt.Sprintf(`{"type":%q,"timestamp":%q}`, typ, ts)
	}
	return fmt.Sprintf(`{"type":%q,"model":%q,"timestamp":%q}`, typ, model, ts)
}

// calRepeat returns n lines built by f.
func calRepeat(n int, f func() string) []string {
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, f())
	}
	return out
}

// buildCalCmd returns the `go build` command for the calibration binary.
func buildCalCmd(t *testing.T, bin string) *exec.Cmd {
	t.Helper()
	c := exec.Command("go", "build", "-o", bin, "./cmd/centinela")
	c.Dir = repoRoot(t)
	return c
}

// idxBefore asserts a appears before b in out (both must be present).
func idxBefore(t *testing.T, out, a, b string) {
	t.Helper()
	ia, ib := strings.Index(out, a), strings.Index(out, b)
	if ia < 0 || ib < 0 {
		t.Fatalf("missing %q(%d) or %q(%d) in:\n%s", a, ia, b, ib, out)
	}
	if ia > ib {
		t.Fatalf("%q should appear before %q:\n%s", a, b, out)
	}
}

// recordSection returns the text block describing model id (renderer joins blocks
// with a blank line; this isolates id's block up to the next blank line).
func recordSection(out, id string) string { return sectionBody(out, id) }
