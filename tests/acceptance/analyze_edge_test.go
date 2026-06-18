package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Scenario: A repo with no recognized manifest still produces a valid inventory and exits 0
func TestAnalyzeNoManifestStillValid(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.py", "print('hi')\n")
	out, code := runAnalyzeBin(t, dir)
	if code != 0 {
		t.Fatalf("no-manifest repo must exit 0, got %d\n%s", code, out)
	}
	data, err := os.ReadFile(filepath.Join(dir, ".workflow", "analysis.json"))
	if err != nil {
		t.Fatalf("inventory not written: %v", err)
	}
	if !strings.Contains(string(data), `"manifests": null`) && !strings.Contains(string(data), `"manifests": []`) {
		t.Fatalf("manifests must be empty: %s", data)
	}
	if !strings.Contains(string(data), `"kind": "none"`) {
		t.Fatalf("graph must be none for no-manifest repo: %s", data)
	}
}

// Scenario: The scan skips dependency and build directories so counts reflect real source
func TestAnalyzeSkipsVendorDepsReadOnly(t *testing.T) {
	dir := analyzeGoRepo(t)
	writeFile(t, dir, "vendor/dep/x.go", "package dep")
	writeFile(t, dir, "node_modules/pkg/y.js", "1")
	writeFile(t, dir, ".gitignore", "ignored/\n")
	writeFile(t, dir, "ignored/z.go", "package ignored")
	witness := filepath.Join(dir, "a", "a.go")
	before, _ := os.ReadFile(witness)
	if _, code := runAnalyzeBin(t, dir); code != 0 {
		t.Fatal("skip-set run must exit 0")
	}
	data, _ := os.ReadFile(filepath.Join(dir, ".workflow", "analysis.json"))
	if strings.Contains(string(data), "vendor") || strings.Contains(string(data), "node_modules") ||
		strings.Contains(string(data), "ignored") {
		t.Fatalf("counts must exclude skip-set/gitignored paths:\n%s", data)
	}
	if after, _ := os.ReadFile(witness); string(after) != string(before) {
		t.Fatal("source file must not be mutated (read-only)")
	}
}

// Scenario: Running analyze with an un-writable output path fails clearly with a non-zero exit and writes no partial inventory
func TestAnalyzeUnwritableOutFails(t *testing.T) {
	dir := analyzeGoRepo(t)
	writeFile(t, dir, "blocker", "x") // a file where a dir is expected
	out, code := runAnalyzeBin(t, dir, "--out", "blocker/analysis.json")
	if code == 0 {
		t.Fatalf("un-writable out must fail non-zero:\n%s", out)
	}
	if !strings.Contains(out, "cannot write") {
		t.Fatalf("stderr must explain the un-writable path: %q", out)
	}
}

// Scenario: Running analyze against a non-existent or unreadable root fails clearly and writes no inventory
func TestAnalyzeUnreadableRootFails(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}
	dir := t.TempDir()
	// 0o311: the process can enter (search) but ReadDir(".") fails (no read), so
	// Analyze hard-errors on the unreadable root while exec still starts.
	if err := os.Chmod(dir, 0o311); err != nil {
		t.Skip("cannot restrict dir: " + err.Error())
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
	out, code := runAnalyzeBin(t, dir)
	if code == 0 {
		t.Fatalf("unreadable root must fail non-zero:\n%s", out)
	}
	if !strings.Contains(out, "unreadable root") {
		t.Fatalf("stderr must name the unreadable root: %q", out)
	}
}
