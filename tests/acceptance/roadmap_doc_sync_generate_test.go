package acceptance_test

import (
	"bytes"
	"strings"
	"testing"
)

// Acceptance: specs/roadmap-doc-sync.feature

// Scenario: generate writes ROADMAP.md and exits 0
func TestRds_GenerateWritesAndExitsZero(t *testing.T) {
	bin := buildCent(t)
	dir := rdsDir(t, sampleRoadmap, rdsToml("warn"))
	out, code := runCent(t, bin, dir, "roadmap", "generate")
	if code != 0 {
		t.Fatalf("exit=%d\n%s", code, out)
	}
	mustHave(t, out, "ROADMAP.md")
	if len(rdsGenerate(t, bin, dir)) == 0 {
		t.Fatal("ROADMAP.md is empty")
	}
}

// Scenario: Generated output is deterministic — running generate twice yields byte-identical files
func TestRds_GenerateDeterministic(t *testing.T) {
	bin := buildCent(t)
	dir := rdsDir(t, sampleRoadmap, rdsToml("warn"))
	first := rdsGenerate(t, bin, dir)
	second := rdsGenerate(t, bin, dir)
	if !bytes.Equal(first, second) {
		t.Fatalf("non-deterministic:\n%q\n%q", first, second)
	}
}

// Scenario: generate creates ROADMAP.md from scratch when the file is absent
func TestRds_GenerateCreatesFromScratch(t *testing.T) {
	bin := buildCent(t)
	dir := rdsDir(t, sampleRoadmap, rdsToml("warn"))
	data := rdsGenerate(t, bin, dir)
	mustHave(t, string(data), "# Roadmap")
}

// Scenario: Generated file ends with exactly one trailing newline and no trailing whitespace
func TestRds_TrailingNewlineAndNoTrailingWS(t *testing.T) {
	bin := buildCent(t)
	dir := rdsDir(t, sampleRoadmap, rdsToml("warn"))
	out := string(rdsGenerate(t, bin, dir))
	if !strings.HasSuffix(out, "\n") || strings.HasSuffix(out, "\n\n") {
		t.Fatalf("must end with exactly one newline: %q", out)
	}
	for _, l := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if l != strings.TrimRight(l, " \t") {
			t.Fatalf("trailing whitespace on %q", l)
		}
	}
}

// Scenario: Generated file uses LF line endings on all platforms
func TestRds_LFLineEndings(t *testing.T) {
	bin := buildCent(t)
	dir := rdsDir(t, sampleRoadmap, rdsToml("warn"))
	if bytes.Contains(rdsGenerate(t, bin, dir), []byte("\r")) {
		t.Fatal("output must be LF-only, found CR")
	}
}
