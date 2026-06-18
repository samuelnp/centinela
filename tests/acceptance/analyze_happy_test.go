package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/deep-codebase-analysis.feature

// Scenario: Analyzing a Go module writes a complete inventory and prints a summary
func TestAnalyzeScanWritesInventory(t *testing.T) {
	dir := analyzeGoRepo(t)
	out, code := runAnalyzeBin(t, dir)
	if code != 0 {
		t.Fatalf("analyze must exit 0, got %d\n%s", code, out)
	}
	if !strings.Contains(out, "primary language: Go") || !strings.Contains(out, "graph edges:") {
		t.Fatalf("summary must report primary + signals: %q", out)
	}
	data, err := os.ReadFile(filepath.Join(dir, ".workflow", "analysis.json"))
	if err != nil {
		t.Fatalf("inventory not written: %v", err)
	}
	for _, want := range []string{`"schemaVersion": 1`, `"primaryLanguage": "Go"`, `"kind": "go-packages"`} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("inventory missing %q:\n%s", want, data)
		}
	}
}

// Scenario: Re-running analyze on an unchanged repo produces a byte-identical inventory
func TestAnalyzeDeterministicRerun(t *testing.T) {
	dir := analyzeGoRepo(t)
	p := filepath.Join(dir, ".workflow", "analysis.json")
	if _, code := runAnalyzeBin(t, dir); code != 0 {
		t.Fatal("first run failed")
	}
	first, _ := os.ReadFile(p)
	if _, code := runAnalyzeBin(t, dir); code != 0 {
		t.Fatal("second run failed")
	}
	second, _ := os.ReadFile(p)
	if string(first) != string(second) {
		t.Fatal("re-run on unchanged repo must be byte-identical")
	}
}

// Scenario: The --out flag redirects the inventory to a custom path
func TestAnalyzeOutOverride(t *testing.T) {
	dir := analyzeGoRepo(t)
	_, code := runAnalyzeBin(t, dir, "--out", "build/inventory.json")
	if code != 0 {
		t.Fatalf("--out run must exit 0, got %d", code)
	}
	if _, err := os.Stat(filepath.Join(dir, "build", "inventory.json")); err != nil {
		t.Fatalf("--out target not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".workflow", "analysis.json")); !os.IsNotExist(err) {
		t.Fatalf("default path must not be written under --out: %v", err)
	}
}
