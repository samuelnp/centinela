package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/analyze"
	"github.com/samuelnp/centinela/internal/reconstruct"
)

// TestReconstructPipeline exercises the full analyze -> Save -> Load ->
// Reconstruct -> WriteCorpus chain on a real on-disk Go n-tier fixture, crossing
// the file-system boundary between the analyze and reconstruct packages and
// asserting the review-dir corpus lands and re-runs byte-identically.
func TestReconstructPipeline(t *testing.T) {
	dir := t.TempDir()
	mk(t, dir, "go.mod", "module fixturemod\n\ngo 1.21\n")
	mk(t, dir, "internal/handler/h.go", "package handler\n")
	mk(t, dir, "internal/service/s.go", "package service\n")
	mk(t, dir, "cmd/app/main.go", "package main\n\nfunc main() {}\n")
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
	r := reconstruct.NewReconstructor().Reconstruct(loaded)
	if len(r.Targets) == 0 {
		t.Fatal("expected at least one reconstructed target for the n-tier fixture")
	}
	out := filepath.Join(dir, "review")
	written, _, err := reconstruct.WriteCorpus(out, r)
	if err != nil {
		t.Fatal(err)
	}
	if len(written) == 0 {
		t.Fatal("expected files written to the review dir")
	}
	first, _ := os.ReadFile(written[0])
	if !strings.Contains(string(first), "Feature:") || !strings.Contains(string(first), "# TODO: confirm") {
		t.Fatalf("reconstructed feature missing Feature:/TODO markers:\n%s", first)
	}
	again, _, err := reconstruct.WriteCorpus(out, r)
	if err != nil {
		t.Fatal(err)
	}
	second, _ := os.ReadFile(again[0])
	if string(first) != string(second) {
		t.Fatal("WriteCorpus must be byte-identical across re-runs")
	}
}
