package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/capability-calibration.feature

// Scenario: Multiple models in one log each receive an independent classification in a single pass
func TestCalMultiModelIndependent(t *testing.T) {
	lines := calConcat(
		calRepeat(4, func() string { return adv("claude-opus-4-7") }), []string{gf("claude-opus-4-7")},
		calRepeat(3, func() string { return adv("claude-sonnet-4-6") }),
		calRepeat(3, func() string { return gf("claude-sonnet-4-6") }),
		calRepeat(2, func() string { return adv("claude-haiku-4-5") }), []string{gf("claude-haiku-4-5")})
	out, code := runCal(t, calRepo(t, lines))
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if !strings.Contains(recordSection(out, "claude-opus-4-7"), "WellCalibrated") {
		t.Fatalf("opus should be WellCalibrated:\n%s", out)
	}
	if !strings.Contains(recordSection(out, "claude-sonnet-4-6"), "Undergoverned") {
		t.Fatalf("sonnet should be Undergoverned:\n%s", out)
	}
	if !strings.Contains(recordSection(out, "claude-haiku-4-5"), "WellCalibrated") {
		t.Fatalf("haiku (2 advances) should be WellCalibrated:\n%s", out)
	}
}

// tieToml maps two arbitrary model ids to the same class/profile for tie-break.
const tieToml = calToml + `"zeta-model" = "capable"
"alpha-model" = "capable"
`

// Scenario: Model id tie-breaking sorts by model id ascending for fully stable ordering
func TestCalTieBreakById(t *testing.T) {
	dir := t.TempDir()
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	if err := os.WriteFile(filepath.Join(dir, "centinela.toml"), []byte(tieToml), 0o644); err != nil {
		t.Fatal(err)
	}
	td := filepath.Join(dir, ".workflow", "telemetry")
	if err := os.MkdirAll(td, 0o755); err != nil {
		t.Fatal(err)
	}
	lines := calConcat(
		calRepeat(3, func() string { return adv("zeta-model") }), []string{gf("zeta-model")},
		calRepeat(3, func() string { return adv("alpha-model") }), []string{gf("alpha-model")})
	if err := os.WriteFile(filepath.Join(td, "events.jsonl"), []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, code := runCent(t, buildCalBin(t), dir, "calibrate")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	idxBefore(t, out, "alpha-model", "zeta-model")
}
