package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/analyze"
)

// analyzeRepo chdirs into a temp Go module with a Makefile, reverting on cleanup.
func analyzeRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	write := func(rel, body string) {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module fixturemod\n\ngo 1.21\n")
	write("main.go", "package main\n\nfunc main() {}\n")
	write("Makefile", "test:\n\tgo test ./...\n")
	o, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(o) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	return dir
}

// runAnalyzeCmd invokes runAnalyze with a fresh buffered command.
func runAnalyzeCmd(t *testing.T, out string) (string, error) {
	t.Helper()
	old := analyzeOut
	analyzeOut = out
	t.Cleanup(func() { analyzeOut = old })
	c := &cobra.Command{}
	var buf bytes.Buffer
	c.SetOut(&buf)
	err := runAnalyze(c, nil)
	return buf.String(), err
}

func TestRunAnalyze_HappyPath(t *testing.T) {
	analyzeRepo(t)
	out, err := runAnalyzeCmd(t, analyze.DefaultOutPath)
	if err != nil {
		t.Fatalf("analyze errored: %v", err)
	}
	if !strings.Contains(out, "primary language: Go") {
		t.Fatalf("summary missing primary language: %q", out)
	}
	if _, err := os.Stat(analyze.DefaultOutPath); err != nil {
		t.Fatalf("inventory not written: %v", err)
	}
}

func TestRunAnalyze_OutOverride(t *testing.T) {
	analyzeRepo(t)
	if _, err := runAnalyzeCmd(t, "build/inv.json"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat("build/inv.json"); err != nil {
		t.Fatalf("--out target not written: %v", err)
	}
	if _, err := os.Stat(analyze.DefaultOutPath); !os.IsNotExist(err) {
		t.Fatalf("default path must not be written under --out: %v", err)
	}
}
