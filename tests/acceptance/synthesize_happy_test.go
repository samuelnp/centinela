package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/archetype-inference-project-synthesis.feature

// Scenario: A Rails-shaped inventory infers rails-native and drafts PROJECT.md
func TestAccSynth_Rails(t *testing.T) {
	dir := t.TempDir()
	in := writeAnalysis(t, dir, railsInventory)
	out, code := runSynthesizeBin(t, dir, "--in", in, "--out", filepath.Join(dir, "PROJECT.md"))
	if code != 0 || !strings.Contains(out, "rails-native") {
		t.Fatalf("expected rails-native (code %d):\n%s", code, out)
	}
	body, _ := os.ReadFile(filepath.Join(dir, "PROJECT.md"))
	if !strings.Contains(string(body), "**Archetype:** rails-native") || !strings.Contains(string(body), "Model") {
		t.Fatalf("rails draft missing archetype/layer mapping:\n%s", body)
	}
}

// Scenario: A Go n-tier inventory infers n-tier with handler service repository mapping
func TestAccSynth_GoNTierEndToEnd(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module fixturemod\n\ngo 1.21\n")
	writeFile(t, dir, "internal/handler/h.go", "package handler\n")
	writeFile(t, dir, "internal/service/s.go", "package service\n")
	writeFile(t, dir, "internal/repository/r.go", "package repository\n")
	if _, code := runAnalyzeBin(t, dir, "--out", filepath.Join(dir, "analysis.json")); code != 0 {
		t.Fatalf("analyze failed (code %d)", code)
	}
	out, code := runSynthesizeBin(t, dir, "--in", filepath.Join(dir, "analysis.json"), "--out", filepath.Join(dir, "P.md"))
	if code != 0 || !strings.Contains(out, "n-tier") {
		t.Fatalf("expected n-tier end-to-end (code %d):\n%s", code, out)
	}
}

// Scenario: A game inventory with systems and components infers ecs
func TestAccSynth_ECS(t *testing.T) {
	dir := t.TempDir()
	in := writeAnalysis(t, dir, ecsInventory)
	out, code := runSynthesizeBin(t, dir, "--in", in, "--out", filepath.Join(dir, "PROJECT.md"))
	if code != 0 || !strings.Contains(out, "ecs") {
		t.Fatalf("expected ecs (code %d):\n%s", code, out)
	}
}

// Scenario: Re-running synthesize on the same inventory is byte-identical
func TestAccSynth_Deterministic(t *testing.T) {
	dir := t.TempDir()
	in := writeAnalysis(t, dir, railsInventory)
	runSynthesizeBin(t, dir, "--in", in, "--out", filepath.Join(dir, "a.md"))
	runSynthesizeBin(t, dir, "--in", in, "--out", filepath.Join(dir, "b.md"))
	a, _ := os.ReadFile(filepath.Join(dir, "a.md"))
	b, _ := os.ReadFile(filepath.Join(dir, "b.md"))
	if string(a) != string(b) || len(a) == 0 {
		t.Fatal("synthesize output must be byte-identical across runs")
	}
}
