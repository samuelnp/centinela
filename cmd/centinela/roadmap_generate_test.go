package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

func seedGenerate(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.MkdirAll(filepath.Join(dir, ".workflow"), 0o755); err != nil {
		t.Fatal(err)
	}
	js := `{"intro":"hi","phases":[{"name":"P","features":[{"name":"a","description":"d"}]}]}`
	if err := os.WriteFile(filepath.Join(dir, ".workflow", "roadmap.json"), []byte(js), 0o644); err != nil {
		t.Fatal(err)
	}
}

// generate writes ROADMAP.md from scratch when absent, byte-equal to RenderMarkdown.
func TestRunRoadmapGenerateCreatesFile(t *testing.T) {
	seedGenerate(t)
	if err := runRoadmapGenerate(nil, nil); err != nil {
		t.Fatalf("generate: %v", err)
	}
	got, err := os.ReadFile(roadmapMarkdownFile)
	if err != nil {
		t.Fatalf("ROADMAP.md not written: %v", err)
	}
	r, _ := roadmap.Load()
	want := roadmap.RenderMarkdown(r)
	if string(got) != string(want) {
		t.Fatalf("file mismatch\n got:%q\nwant:%q", got, want)
	}
}

// generate is idempotent: running twice yields byte-identical output.
func TestRunRoadmapGenerateIdempotent(t *testing.T) {
	seedGenerate(t)
	if err := runRoadmapGenerate(nil, nil); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(roadmapMarkdownFile)
	if err := runRoadmapGenerate(nil, nil); err != nil {
		t.Fatal(err)
	}
	second, _ := os.ReadFile(roadmapMarkdownFile)
	if string(first) != string(second) {
		t.Fatalf("not idempotent: %q vs %q", first, second)
	}
}

// A missing roadmap.json is surfaced as a command error, not a panic.
func TestRunRoadmapGenerateLoadError(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := runRoadmapGenerate(nil, nil); err == nil {
		t.Fatal("expected error when roadmap.json is absent")
	}
}

// A write failure (ROADMAP.md is a directory) is surfaced as an error.
func TestRunRoadmapGenerateWriteError(t *testing.T) {
	seedGenerate(t)
	if err := os.Mkdir(roadmapMarkdownFile, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := runRoadmapGenerate(nil, nil); err == nil {
		t.Fatal("expected write error when ROADMAP.md is a directory")
	}
}
