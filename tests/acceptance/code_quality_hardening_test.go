package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/code-quality-hardening.feature

func scriptPath(t *testing.T) string {
	t.Helper()
	p := filepath.Join("..", "..", "scripts", "check-fmt.sh")
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("check-fmt.sh missing: %v", err)
	}
	return p
}

// Scenario: Unformatted Go source fails the format check.
func TestUnformattedSourceFailsFormatCheck(t *testing.T) {
	script := scriptPath(t)
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.go")
	// Mis-indented body — gofmt -l will flag this file.
	if err := os.WriteFile(bad, []byte("package x\nfunc F(){\nreturn\n}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("sh", script)
	cmd.Env = append(os.Environ(), "FMT_DIRS="+dir)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected non-zero exit for unformatted file, got success: %s", out)
	}
	if !strings.Contains(string(out), "bad.go") {
		t.Fatalf("output must name the offending file, got: %s", out)
	}
}

// Scenario: Formatted tree passes the format check (exit 0, no output).
func TestFormattedTreePassesFormatCheck(t *testing.T) {
	script := scriptPath(t)
	dir := t.TempDir()
	good := filepath.Join(dir, "good.go")
	if err := os.WriteFile(good, []byte("package x\n\nfunc F() {\n\treturn\n}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("sh", script)
	cmd.Env = append(os.Environ(), "FMT_DIRS="+dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("formatted tree must pass, got error: %v (%s)", err, out)
	}
	if len(strings.TrimSpace(string(out))) != 0 {
		t.Fatalf("formatted tree must produce no output, got: %s", out)
	}
}

// Scenario: Validate suite gates formatting.
func TestValidateSuiteGatesFormatting(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "centinela.toml"))
	if err != nil {
		t.Fatalf("read centinela.toml: %v", err)
	}
	if !strings.Contains(string(data), "./scripts/check-fmt.sh") {
		t.Fatal("[validate] commands must include ./scripts/check-fmt.sh")
	}
}
