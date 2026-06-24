package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/analyze"
	"github.com/samuelnp/centinela/internal/brownmap"
)

// TestBrownfieldPipeline runs the full analyze -> Save -> Load -> Generate ->
// WriteDraft chain on a real on-disk Go n-tier fixture, crossing the file-system
// boundary, and asserts that a pre-existing canonical roadmap.json is left
// byte-for-byte unchanged by the brownfield run.
func TestBrownfieldPipeline(t *testing.T) {
	dir := t.TempDir()
	mk(t, dir, "go.mod", "module fixturemod\n\ngo 1.21\n")
	mk(t, dir, "internal/handler/h.go", "package handler\n")
	mk(t, dir, "internal/service/s.go", "package service\n")
	mk(t, dir, "cmd/app/main.go", "package main\n\nfunc main() {}\n")
	const curated = `{"phases":[{"name":"Phase 1","features":[{"name":"hand-authored"}]}]}`
	mk(t, dir, ".workflow/roadmap.json", curated)
	chdir(t, dir)

	inv, err := analyze.Analyze(".")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "analysis.json")
	if err := analyze.Save(path, inv); err != nil {
		t.Fatal(err)
	}
	loaded, err := analyze.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	plan := brownmap.NewBrownfielder().Generate(loaded, []string{"Add OAuth login"})
	if plan.BaselineCount == 0 {
		t.Fatal("expected a non-empty Baseline for the n-tier fixture")
	}
	out := filepath.Join(dir, brownmap.DefaultDraftPath)
	if _, err := brownmap.WriteDraft(out, plan); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("draft not written: %v", err)
	}
	after, _ := os.ReadFile(filepath.Join(dir, ".workflow", "roadmap.json"))
	if string(after) != curated {
		t.Fatal("existing roadmap.json must be byte-unchanged after the brownfield run")
	}
}
