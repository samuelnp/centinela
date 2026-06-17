package acceptance_test

import (
	"strings"
	"testing"
)

// Acceptance: specs/capability-calibration.feature

// Scenario: Rate exactly equal to highFrictionRate (1.0) triggers Undergoverned classification
func TestCalBoundaryHighInclusive(t *testing.T) {
	lines := calConcat(calRepeat(3, func() string { return adv("claude-sonnet-4-6") }),
		calRepeat(3, func() string { return gf("claude-sonnet-4-6") })) // rate 1.0
	out, code := runCal(t, calRepo(t, lines))
	sec := recordSection(out, "claude-sonnet-4-6")
	if code != 0 || !strings.Contains(sec, "Undergoverned") || !strings.Contains(sec, "Tighten") {
		t.Fatalf("rate==1.0 should be Undergoverned/Tighten (code %d):\n%s", code, sec)
	}
}

// Scenario: Rate exactly equal to lowFrictionRate (0.25) triggers Overgoverned classification
func TestCalBoundaryLowInclusive(t *testing.T) {
	lines := calConcat(calRepeat(4, func() string { return adv("claude-haiku-4-5") }),
		[]string{gf("claude-haiku-4-5")}) // rate 0.25
	out, code := runCal(t, calRepo(t, lines))
	sec := recordSection(out, "claude-haiku-4-5")
	if code != 0 || !strings.Contains(sec, "Overgoverned") || !strings.Contains(sec, "Loosen") {
		t.Fatalf("rate==0.25 should be Overgoverned/Loosen (code %d):\n%s", code, sec)
	}
}

// Scenario: Model with advances and zero rework has Rate 0.0 and is classified Overgoverned if loosenable
func TestCalBoundaryZeroRateLoosen(t *testing.T) {
	lines := calRepeat(5, func() string { return adv("claude-haiku-4-5") })
	out, code := runCal(t, calRepo(t, lines))
	sec := recordSection(out, "claude-haiku-4-5")
	for _, w := range []string{"advances=5", "rework=0", "rate=0.00", "Overgoverned", "Loosen"} {
		if code != 0 || !strings.Contains(sec, w) {
			t.Fatalf("rate==0.0 missing %q (code %d):\n%s", w, code, sec)
		}
	}
}

// Scenario: Model with only rework events and zero advances is WellCalibrated not Undergoverned
func TestCalBoundaryOnlyReworkWellCalibrated(t *testing.T) {
	lines := calRepeat(10, func() string { return gf("claude-sonnet-4-6") })
	out, code := runCal(t, calRepo(t, lines))
	sec := recordSection(out, "claude-sonnet-4-6")
	for _, w := range []string{"advances=0", "rate=n/a", "WellCalibrated", "Keep"} {
		if code != 0 || !strings.Contains(sec, w) {
			t.Fatalf("only-rework missing %q (code %d):\n%s", w, code, sec)
		}
	}
}
