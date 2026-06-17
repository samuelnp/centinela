package acceptance_test

import (
	"strings"
	"testing"
)

// Acceptance: specs/capability-calibration.feature

// calConcat flattens slices of JSONL lines into one slice.
func calConcat(groups ...[]string) []string {
	var out []string
	for _, g := range groups {
		out = append(out, g...)
	}
	return out
}

// Scenario: Model with high friction under a tightenable profile is classified Undergoverned and recommended tighter profile
func TestCalUndergovernedTighten(t *testing.T) {
	lines := calConcat(calRepeat(3, func() string { return adv("claude-sonnet-4-6") }),
		calRepeat(2, func() string { return gf("claude-sonnet-4-6") }), []string{vr("claude-sonnet-4-6")})
	out, code := runCal(t, calRepo(t, lines))
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	sec := recordSection(out, "claude-sonnet-4-6")
	for _, w := range []string{"Undergoverned", "strict", "Tighten", "advances=3", "rework=3", "rate=1.00"} {
		if !strings.Contains(sec, w) {
			t.Fatalf("missing %q in:\n%s", w, sec)
		}
	}
}

// Scenario: Model with low friction under a loosenable profile is classified Overgoverned and recommended looser profile
func TestCalOvergovernedLoosen(t *testing.T) {
	lines := calConcat(calRepeat(4, func() string { return adv("claude-haiku-4-5") }), []string{gf("claude-haiku-4-5")})
	out, code := runCal(t, calRepo(t, lines))
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	sec := recordSection(out, "claude-haiku-4-5")
	for _, w := range []string{"Overgoverned", "guided", "Loosen", "advances=4", "rework=1", "rate=0.25"} {
		if !strings.Contains(sec, w) {
			t.Fatalf("missing %q in:\n%s", w, sec)
		}
	}
}

// Scenario: Model already at the strictest profile but high friction is classified WellCalibrated with recommendation Keep
func TestCalMaxedStrictKeep(t *testing.T) {
	lines := calConcat(calRepeat(3, func() string { return adv("claude-haiku-4-5") }),
		calRepeat(3, func() string { return gf("claude-haiku-4-5") }))
	out, code := runCal(t, calRepo(t, lines))
	sec := recordSection(out, "claude-haiku-4-5")
	if code != 0 || !strings.Contains(sec, "WellCalibrated") || !strings.Contains(sec, "Keep") {
		t.Fatalf("maxed-strict wrong (code %d):\n%s", code, sec)
	}
}

// Scenario: Model already at the loosest profile but low friction is classified WellCalibrated with recommendation Keep
func TestCalMaxedOutcomeKeep(t *testing.T) {
	lines := calConcat(calRepeat(5, func() string { return adv("claude-opus-4-7") }), []string{gf("claude-opus-4-7")})
	out, code := runCal(t, calRepo(t, lines))
	sec := recordSection(out, "claude-opus-4-7")
	if code != 0 || !strings.Contains(sec, "WellCalibrated") || !strings.Contains(sec, "Keep") {
		t.Fatalf("maxed-outcome wrong (code %d):\n%s", code, sec)
	}
}

// Scenario: Model with friction between thresholds is classified WellCalibrated and recommended Keep
func TestCalBetweenKeep(t *testing.T) {
	lines := calConcat(calRepeat(4, func() string { return adv("claude-sonnet-4-6") }),
		calRepeat(2, func() string { return gf("claude-sonnet-4-6") }))
	out, code := runCal(t, calRepo(t, lines))
	sec := recordSection(out, "claude-sonnet-4-6")
	for _, w := range []string{"WellCalibrated", "Keep", "advances=4", "rework=2", "rate=0.50"} {
		if code != 0 || !strings.Contains(sec, w) {
			t.Fatalf("between wrong (code %d) missing %q:\n%s", code, w, sec)
		}
	}
}
