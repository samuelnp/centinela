// Acceptance: specs/guided-by-default.feature
package acceptance_test

import (
	"path/filepath"
	"strings"
	"testing"
)

// Scenario: A strict greenfield project still requires the full cascade
//
// The cross-profile hole: seeding made the guided cold walk traversable, but the
// seeded artifacts carry the evaluator ROLE, so one guided `roadmap promote`
// followed by pinning strict used to satisfy the strict grading rung with no
// senior-PM or quality-evaluator pass ever having run. Seeded artifacts are now
// marked provisional and strict refuses them.
func TestGBD_GuidedSeedCannotSatisfyTheStrictCascade(t *testing.T) {
	bin := buildCent(t)
	dir := t.TempDir()
	gbdWalkTree(t, dir)

	if out, code := runCent(t, bin, dir, "roadmap", "promote", "finding",
		"--phase", "Phase 0: Bootstrap", "--scores", "1,1,1,1,1,1"); code != 0 {
		t.Fatalf("the guided promote must still succeed: %s", out)
	}
	// Round 1's guarantee must survive: the guided walk is still traversable.
	if out, code := runCent(t, bin, dir, "start", "finding"); code != 0 {
		t.Fatalf("guided start must still succeed after promote: %s", out)
	}

	// Now pin strict. The seeded artifacts must NOT pass for an evaluation.
	mustWrite(t, filepath.Join(dir, "centinela.toml"),
		"[workflow]\nenforcement_profile = \"strict\"\n")
	out, code := runCent(t, bin, dir, "start", "setup")
	if code == 0 {
		t.Fatalf("strict must not accept seeded artifacts as an evaluator pass: %s", out)
	}
	if !strings.Contains(out, "provisional") {
		t.Fatalf("the refusal must say the artifacts are provisional: %s", out)
	}
	if !strings.Contains(out, "roadmap-analysis.json") {
		t.Fatalf("the refusal must name the file to fix: %s", out)
	}
}

// Scenario: A strict greenfield project still requires the full cascade
//
// The ✅ counterweight: clearing the mark is what a real evaluator pass does, and
// it must actually unblock strict. Otherwise the refusal is a dead end rather
// than a redirect.
func TestGBD_ClearingProvisionalUnblocksStrict(t *testing.T) {
	bin := buildCent(t)
	dir := t.TempDir()
	gbdWalkTree(t, dir)
	if out, code := runCent(t, bin, dir, "roadmap", "promote", "finding",
		"--phase", "Phase 0: Bootstrap", "--scores", "9,9,9,9,9,9"); code != 0 {
		t.Fatalf("guided promote: %s", out)
	}
	mustWrite(t, filepath.Join(dir, "centinela.toml"),
		"[workflow]\nenforcement_profile = \"strict\"\n")
	for _, name := range []string{"roadmap-analysis.json", "roadmap-quality.json"} {
		path := filepath.Join(dir, ".workflow", name)
		mustWrite(t, path, strings.Replace(readFile(t, path), "\"provisional\": true,\n", "", 1))
	}
	// The seeded pair now covers only "finding"; grade the rest as an evaluator
	// would, then strict must proceed on its own merits.
	if out, code := runCent(t, bin, dir, "start", "finding"); code != 0 &&
		!strings.Contains(out, "missing feature") {
		t.Fatalf("strict must fail on real coverage, not on the mark: %s", out)
	}
	if out, _ := runCent(t, bin, dir, "start", "finding"); strings.Contains(out, "provisional") {
		t.Fatalf("the provisional refusal must be gone once the mark is cleared: %s", out)
	}
}
