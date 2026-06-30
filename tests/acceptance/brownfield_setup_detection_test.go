package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/brownfield-setup-detection.feature

// runSetupHook runs `centinela hook setup` in dir with an empty JSON event on
// stdin (mirroring the UserPromptSubmit hook), returning combined output + code.
func runSetupHook(t *testing.T, dir string) (string, int) {
	t.Helper()
	c := exec.Command(buildAnalyzeBin(t), "hook", "setup")
	c.Dir = dir
	c.Stdin = strings.NewReader("{}")
	out, err := c.CombinedOutput()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("run hook setup: %v", err)
	}
	return string(out), code
}

// setupRepo makes an initialized repo (centinela.toml, no PROJECT.md) then
// writes each rel path in files (a trailing-slash entry makes an empty dir).
func setupRepo(t *testing.T, files ...string) string {
	t.Helper()
	dir := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(dir, rel)
		os.MkdirAll(filepath.Dir(p), 0o755) //nolint:errcheck
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("centinela.toml", "x")
	for _, f := range files {
		if strings.HasSuffix(f, "/") {
			os.MkdirAll(filepath.Join(dir, f), 0o755) //nolint:errcheck
			continue
		}
		write(f, "x")
	}
	return dir
}

// Scenario: Brownfield repo with go.mod emits the brownfield directive
// Scenario: Brownfield repo with package.json emits the brownfield directive
// Scenario: Brownfield repo with a populated src/ directory is detected as brownfield
// Scenario: Brownfield repo with populated internal/ directory is detected as brownfield
func TestAccBrownfield_SourceSignalsRouteBrownfield(t *testing.T) {
	for _, sig := range []string{"go.mod", "package.json", "src/main.go", "internal/x.go"} {
		out, code := runSetupHook(t, setupRepo(t, sig))
		if code != 0 {
			t.Fatalf("%s: expected exit 0, got %d:\n%s", sig, code, out)
		}
		if !strings.Contains(out, "BROWNFIELD PROJECT DETECTED") ||
			!strings.Contains(out, "centinela analyze") ||
			!strings.Contains(out, "centinela synthesize") ||
			!strings.Contains(out, "**Project Stage:** existing") {
			t.Fatalf("%s: expected brownfield directive, got:\n%s", sig, out)
		}
	}
}
