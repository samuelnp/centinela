// Acceptance: specs/guided-by-default.feature
package acceptance_test

import (
	"path/filepath"
	"strings"
	"testing"
)

// gbdWalkTree lays the AC6 cold-start tree and adds the two features `start`
// legitimately refuses: a Backlog finding and a draft. Both resolve into the
// bootstrap phase, so the walk stays inside the one phase a cold start may
// schedule. Still no ROADMAP.md, no analysis, no quality report.
func gbdWalkTree(t *testing.T, dir string) {
	t.Helper()
	mustWrite(t, filepath.Join(dir, "PROJECT.md"), "Project Stage: greenfield\n")
	mustWrite(t, filepath.Join(dir, ".workflow", "roadmap.json"),
		`{"phases":[`+
			`{"name":"Phase 0: Bootstrap","features":[{"name":"setup"},{"name":"a-draft","draft":true}]},`+
			`{"name":"Backlog","features":[{"name":"finding","summary":"a deferred finding"}]}`+
			`]}`)
}

// Scenario: A guided greenfield project cold-starts from PROJECT.md and roadmap.json alone
// Scenario: Guided still refuses what the roadmap says it must
//
// Walks the WHOLE advertised path rather than only its first step: the refusal
// `start` gives must name a command that actually runs on this tree, that
// command must succeed, and the feature must then be startable. Asserting only
// the refusal is what let the dead end ship.
func TestGBD_GuidedColdStartBacklogWalkIsTraversable(t *testing.T) {
	bin := buildCent(t)
	dir := t.TempDir()
	gbdWalkTree(t, dir)

	out, code := runCent(t, bin, dir, "start", "finding")
	if code == 0 || !strings.Contains(out, "roadmap promote") {
		t.Fatalf("a Backlog finding must be refused and sent to promote: %s", out)
	}
	out, code = runCent(t, bin, dir, "roadmap", "promote", "finding",
		"--phase", "Phase 0: Bootstrap", "--scores", "3,3,3,3,3,3")
	if code != 0 {
		t.Fatalf("the promote `start` named must run on this very tree: %s", out)
	}
	out, code = runCent(t, bin, dir, "start", "finding")
	if code != 0 {
		t.Fatalf("the promoted feature must now start: %s", out)
	}
	if !strings.Contains(out, "Current step") || !strings.Contains(out, "plan") {
		t.Fatalf("the workflow must land on the plan step: %s", out)
	}
}

// Scenario: Guided still refuses what the roadmap says it must
//
// The draft branch of the same walk.
func TestGBD_GuidedColdStartDraftWalkIsTraversable(t *testing.T) {
	bin := buildCent(t)
	dir := t.TempDir()
	gbdWalkTree(t, dir)

	out, code := runCent(t, bin, dir, "start", "a-draft")
	if code == 0 || !strings.Contains(out, "draft") || !strings.Contains(out, "roadmap promote") {
		t.Fatalf("a draft must be refused, naming \"draft\" and the finalize path: %s", out)
	}
	out, code = runCent(t, bin, dir, "roadmap", "promote", "a-draft", "--scores", "3,3,3,3,3,3")
	if code != 0 {
		t.Fatalf("the finalize `start` named must run on this very tree: %s", out)
	}
	out, code = runCent(t, bin, dir, "start", "a-draft")
	if code != 0 || !strings.Contains(out, "plan") {
		t.Fatalf("the finalized draft must now start on plan: %s", out)
	}
}

// Scenario: A strict greenfield project still requires the full cascade
//
// The ❌ direction for the same walk: under strict, promote must still refuse a
// missing artifact and leave the roadmap untouched.
func TestGBD_StrictColdWalkStillRefusesPromote(t *testing.T) {
	bin := buildCent(t)
	dir := t.TempDir()
	gbdWalkTree(t, dir)
	mustWrite(t, filepath.Join(dir, "centinela.toml"),
		"[workflow]\nenforcement_profile = \"strict\"\n")
	before := readFile(t, filepath.Join(dir, ".workflow", "roadmap.json"))
	out, code := runCent(t, bin, dir, "roadmap", "promote", "finding",
		"--phase", "Phase 0: Bootstrap", "--scores", "9,9,9,9,9,9")
	if code == 0 || !strings.Contains(out, "roadmap artifact json missing") {
		t.Fatalf("strict must still refuse a missing grading artifact: %s", out)
	}
	if readFile(t, filepath.Join(dir, ".workflow", "roadmap.json")) != before {
		t.Fatal("a refused promote must leave roadmap.json byte-identical")
	}
}
