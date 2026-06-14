package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Integration: a full read-only doctor run on a synthesized drifted repo, then
// a --fix round-trip that repairs every safe issue, then an idempotent second
// --fix that changes nothing.
func TestDoctorFullRunAndFixRoundTrip(t *testing.T) {
	bin := buildDoctor(t)
	dir := seedDriftedRepo(t)

	// 1. Full read-only run surfaces all drift as errors (exit 1).
	out, code := runDoc(t, bin, dir)
	if code != 1 {
		t.Fatalf("drifted repo must exit 1, got %d\n%s", code, out)
	}
	for _, name := range []string{"✗ hooks", "✗ roadmap", "✗ evidence"} {
		if !strings.Contains(out, name) {
			t.Fatalf("expected %q in full run:\n%s", name, out)
		}
	}

	// 2. --fix repairs every safe issue → re-diagnosis all OK (exit 0).
	out2, code2 := runDoc(t, bin, dir, "--fix")
	if code2 != 0 {
		t.Fatalf("--fix must clear all errors, got %d\n%s", code2, out2)
	}
	for _, name := range []string{"✓ hooks", "✓ roadmap", "✓ evidence"} {
		if !strings.Contains(out2, name) {
			t.Fatalf("expected %q post-fix:\n%s", name, out2)
		}
	}
	js, _ := os.ReadFile(filepath.Join(dir, ".workflow", "roadmap.json"))
	if strings.Contains(string(js), "✅") {
		t.Fatal("glyph must be stripped from roadmap.json")
	}
	if left, _ := filepath.Glob(filepath.Join(dir, ".workflow", "*.json.tmp")); len(left) != 0 {
		t.Fatalf("tmp must be swept, left %v", left)
	}

	// 3. Second --fix is a no-op: state byte-identical, still exit 0.
	jsBefore, _ := os.ReadFile(filepath.Join(dir, ".workflow", "roadmap.json"))
	mdBefore, _ := os.ReadFile(filepath.Join(dir, "ROADMAP.md"))
	out3, code3 := runDoc(t, bin, dir, "--fix")
	if code3 != 0 {
		t.Fatalf("idempotent --fix must exit 0, got %d\n%s", code3, out3)
	}
	jsAfter, _ := os.ReadFile(filepath.Join(dir, ".workflow", "roadmap.json"))
	mdAfter, _ := os.ReadFile(filepath.Join(dir, "ROADMAP.md"))
	if string(jsBefore) != string(jsAfter) || string(mdBefore) != string(mdAfter) {
		t.Fatal("second --fix must leave roadmap.json + ROADMAP.md byte-identical")
	}
}
