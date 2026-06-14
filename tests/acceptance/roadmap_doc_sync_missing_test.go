package acceptance_test

import (
	"strings"
	"testing"
)

// Acceptance: specs/roadmap-doc-sync.feature

// Scenario: Missing ROADMAP.md is reported as a clear failure under severity fail
func TestRds_MissingUnderFail(t *testing.T) {
	bin := buildCent(t)
	dir := rdsDir(t, sampleRoadmap, rdsToml("fail"))
	out, code := rdsValidate(t, bin, dir)
	mustHave(t, out, "missing")
	mustHave(t, out, "roadmap generate")
	if code == 0 {
		t.Fatalf("missing under fail must exit non-zero\n%s", out)
	}
}

// Scenario: Missing ROADMAP.md produces a warn result under severity warn
func TestRds_MissingUnderWarn(t *testing.T) {
	bin := buildCent(t)
	dir := rdsDir(t, sampleRoadmap, rdsToml("warn"))
	out, _ := rdsValidate(t, bin, dir)
	mustHave(t, out, "missing")
	mustHave(t, out, "roadmap generate")
}

// Scenario: Gate disabled skips the check even when ROADMAP.md is absent
func TestRds_GateDisabledSkips(t *testing.T) {
	bin := buildCent(t)
	dir := rdsDir(t, sampleRoadmap, rdsDisabledToml)
	out, code := rdsValidate(t, bin, dir)
	if strings.Contains(out, "roadmap_drift") {
		t.Fatalf("disabled gate must not appear:\n%s", out)
	}
	if code != 0 {
		t.Fatalf("disabled gate validate must exit 0\n%s", out)
	}
}
