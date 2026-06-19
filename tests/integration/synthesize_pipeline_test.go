package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/analyze"
	"github.com/samuelnp/centinela/internal/synthesize"
)

// TestSynthesizePipeline exercises the full analyze -> Save -> Load -> Infer ->
// Draft -> WriteDraft chain on a real on-disk Go n-tier fixture, crossing the
// file-system boundary between the analyze and synthesize packages.
func TestSynthesizePipeline(t *testing.T) {
	dir := t.TempDir()
	mk(t, dir, "go.mod", "module fixturemod\n\ngo 1.21\n")
	mk(t, dir, "internal/handler/h.go", "package handler\n")
	mk(t, dir, "internal/service/s.go", "package service\n")
	mk(t, dir, "internal/repository/r.go", "package repository\n")
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
	inf := synthesize.NewInferer().Infer(loaded)
	if inf.Best != synthesize.NTier {
		t.Fatalf("expected n-tier, got %s (scores %+v)", inf.Best, inf.Scores)
	}
	written, _, err := synthesize.WriteDraft(filepath.Join(dir, "PROJECT.md"), synthesize.Draft(loaded, inf))
	if err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(written)
	if !strings.Contains(string(body), "**Archetype:** n-tier") {
		t.Fatalf("pipeline draft wrong:\n%s", body)
	}
}
