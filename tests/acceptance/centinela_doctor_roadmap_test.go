package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/centinela-doctor.feature

// Scenario: ROADMAP.md drift from roadmap.json is flagged
func TestDoctorRoadmapDriftFlagged(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	seedRoadmap(t, dir, "Phase 1: Core")
	writeFile(t, dir, "ROADMAP.md", "# Roadmap\n\nhand-edited\n")
	out, code := runDoctor(t, dir)
	if code != 1 || !strings.Contains(out, "✗ roadmap") {
		t.Fatalf("drift must Error/exit 1, got %d\n%s", code, out)
	}
	if !strings.Contains(out, "out of sync") {
		t.Fatalf("must indicate out-of-sync:\n%s", out)
	}
}

// Scenario: Roadmap drift is repaired by --fix via regeneration
func TestDoctorRoadmapDriftRepaired(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	seedRoadmap(t, dir, "Phase 1: Core")
	writeFile(t, dir, "ROADMAP.md", "stale\n")
	out, code := runDoctor(t, dir, "--fix")
	if code != 0 || strings.Contains(out, "✗ roadmap") {
		t.Fatalf("--fix must regenerate ROADMAP.md, got %d\n%s", code, out)
	}
	md, _ := os.ReadFile(filepath.Join(dir, "ROADMAP.md"))
	if strings.Contains(string(md), "stale") {
		t.Fatal("ROADMAP.md must be regenerated, still stale")
	}
}

// Scenario: Phase name containing a live-status glyph is flagged as ERROR
func TestDoctorRoadmapGlyphFlagged(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	seedRoadmap(t, dir, "✅ Phase 0: Bootstrap")
	out, code := runDoctor(t, dir)
	if code != 1 || !strings.Contains(out, "✗ roadmap") {
		t.Fatalf("glyph must Error/exit 1, got %d\n%s", code, out)
	}
	if !strings.Contains(out, "Phase 0") || !strings.Contains(out, "prefix") {
		t.Fatalf("must name offending phase + prefix breakage:\n%s", out)
	}
}

// Scenario: Phase-name glyph is stripped by --fix and re-diagnosis passes
func TestDoctorRoadmapGlyphStripped(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	seedRoadmap(t, dir, "✅ Phase 0: Bootstrap")
	out, code := runDoctor(t, dir, "--fix")
	if code != 0 || strings.Contains(out, "✗ roadmap") {
		t.Fatalf("--fix must strip glyph, got %d\n%s", code, out)
	}
	js, _ := os.ReadFile(filepath.Join(dir, ".workflow", "roadmap.json"))
	if strings.Contains(string(js), "✅") {
		t.Fatal("glyph must be stripped from roadmap.json")
	}
}

// Scenario: Roadmap glyph strip under --fix is idempotent
func TestDoctorRoadmapGlyphFixIdempotent(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	seedRoadmap(t, dir, "✅ Phase 0: Bootstrap")
	runDoctor(t, dir, "--fix")
	before, _ := os.ReadFile(filepath.Join(dir, ".workflow", "roadmap.json"))
	out, code := runDoctor(t, dir, "--fix")
	after, _ := os.ReadFile(filepath.Join(dir, ".workflow", "roadmap.json"))
	if string(before) != string(after) {
		t.Fatal("second --fix must leave roadmap.json byte-identical")
	}
	if code != 0 || strings.Contains(out, "✗ roadmap") {
		t.Fatalf("idempotent --fix must stay OK, got %d\n%s", code, out)
	}
}
