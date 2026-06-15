package acceptance_test

import (
	"strings"
	"testing"
)

// Acceptance: specs/capability-calibration.feature

// Scenario: Model with fewer than 3 advances is classified WellCalibrated due to insufficient evidence regardless of rate
func TestCalInsufficientAdvances(t *testing.T) {
	lines := calConcat(calRepeat(2, func() string { return adv("claude-haiku-4-5") }),
		calRepeat(5, func() string { return gf("claude-haiku-4-5") }))
	out, code := runCal(t, calRepo(t, lines))
	sec := recordSection(out, "claude-haiku-4-5")
	if code != 0 || !strings.Contains(sec, "WellCalibrated") || !strings.Contains(sec, "Keep") {
		t.Fatalf("insufficient-advances wrong (code %d):\n%s", code, sec)
	}
}

// Scenario: Model with zero step-advanced events is guarded against division-by-zero and classified WellCalibrated
func TestCalZeroAdvancesGuard(t *testing.T) {
	lines := calRepeat(4, func() string { return gf("claude-haiku-4-5") })
	out, code := runCal(t, calRepo(t, lines))
	sec := recordSection(out, "claude-haiku-4-5")
	if code != 0 || !strings.Contains(sec, "advances=0") || !strings.Contains(sec, "rate=n/a") ||
		!strings.Contains(sec, "WellCalibrated") {
		t.Fatalf("zero-advance guard wrong (code %d):\n%s", code, sec)
	}
	if strings.Contains(out, "NaN") || strings.Contains(out, "panic") {
		t.Fatalf("NaN/panic leaked:\n%s", out)
	}
}

// Scenario: Model id with no capability class is classified Unclassified with no recommendation
func TestCalUnclassifiedModel(t *testing.T) {
	lines := calConcat(calRepeat(5, func() string { return adv("local/unknown-model") }),
		calRepeat(5, func() string { return gf("local/unknown-model") }))
	out, code := runCal(t, calRepo(t, lines))
	sec := recordSection(out, "local/unknown-model")
	if code != 0 || !strings.Contains(sec, "Unclassified") || !strings.Contains(sec, "None") {
		t.Fatalf("unclassified wrong (code %d):\n%s", code, sec)
	}
	for _, bad := range []string{"panic", "Error:"} {
		if strings.Contains(out, bad) {
			t.Fatalf("output contains %q:\n%s", bad, out)
		}
	}
}

// Scenario: Unattributed bucket from events with no model is classified Unclassified and rendered last
func TestCalUnattributedLast(t *testing.T) {
	lines := calConcat(
		calRepeat(5, func() string { return adv("") }), calRepeat(5, func() string { return gf("") }),
		calRepeat(3, func() string { return adv("claude-sonnet-4-6") }),
		calRepeat(3, func() string { return gf("claude-sonnet-4-6") }))
	out, code := runCal(t, calRepo(t, lines))
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	sec := recordSection(out, "unattributed")
	if !strings.Contains(sec, "Unclassified") {
		t.Fatalf("unattributed not Unclassified:\n%s", sec)
	}
	idxBefore(t, out, "claude-sonnet-4-6", "unattributed")
}
