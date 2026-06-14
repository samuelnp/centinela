package acceptance_test

import (
	"os"
	"path/filepath"
	"testing"
)

// Acceptance: specs/roadmap-doc-sync.feature

func handEdit(t *testing.T, dir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "ROADMAP.md"), []byte("# Roadmap\n\nhand-edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Scenario: Drift gate passes when ROADMAP.md matches generator output
func TestRds_DriftPassInSync(t *testing.T) {
	bin := buildCent(t)
	dir := rdsDir(t, sampleRoadmap, rdsToml("fail"))
	rdsGenerate(t, bin, dir)
	out, code := rdsValidate(t, bin, dir)
	mustHave(t, out, "roadmap_drift")
	mustHave(t, out, "in sync")
	if code != 0 {
		t.Fatalf("in-sync validate must exit 0\n%s", out)
	}
}

// Scenario: Drift gate fails when ROADMAP.md is hand-edited under severity fail
func TestRds_DriftFailUnderFail(t *testing.T) {
	bin := buildCent(t)
	dir := rdsDir(t, sampleRoadmap, rdsToml("fail"))
	rdsGenerate(t, bin, dir)
	handEdit(t, dir)
	out, code := rdsValidate(t, bin, dir)
	mustHave(t, out, "roadmap_drift")
	mustHave(t, out, "line")
	mustHave(t, out, "roadmap generate")
	if code == 0 {
		t.Fatalf("drift under fail must exit non-zero\n%s", out)
	}
}

// Scenario: Drift gate warns but does not block when ROADMAP.md drifts under severity warn
func TestRds_DriftWarnUnderWarn(t *testing.T) {
	bin := buildCent(t)
	dir := rdsDir(t, sampleRoadmap, rdsToml("warn"))
	rdsGenerate(t, bin, dir)
	handEdit(t, dir)
	out, code := rdsValidate(t, bin, dir)
	mustHave(t, out, "roadmap_drift")
	mustHave(t, out, "roadmap generate")
	if code != 0 {
		t.Fatalf("drift under warn must exit 0\n%s", out)
	}
}

// Scenario: Running generate after a drift failure then re-validating passes the gate
func TestRds_RegenerateThenPasses(t *testing.T) {
	bin := buildCent(t)
	dir := rdsDir(t, sampleRoadmap, rdsToml("fail"))
	rdsGenerate(t, bin, dir)
	handEdit(t, dir)
	if _, code := rdsValidate(t, bin, dir); code == 0 {
		t.Fatal("expected drift fail before regenerate")
	}
	rdsGenerate(t, bin, dir)
	out, code := rdsValidate(t, bin, dir)
	mustHave(t, out, "in sync")
	if code != 0 {
		t.Fatalf("after regenerate validate must exit 0\n%s", out)
	}
}
