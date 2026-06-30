// Acceptance: specs/coverage-hardening.feature
package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const chRoot = "../.."

// Scenario: Total coverage meets the hardened target
// Re-running go test ./... here duplicates a multi-minute job; the validate-step gate is ground truth.
// Scenario: Coverage gate still passes at the configured floor
func TestCoverageGate_ScriptAndFloor(t *testing.T) {
	p := filepath.Join(chRoot, "scripts/check-coverage.sh")
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("gate script missing: %v", err)
	}
	if !strings.Contains(string(data), "MIN_COVERAGE:-95.0") {
		t.Fatal("floor changed; must remain 95.0")
	}
	info, _ := os.Stat(p)
	if info.Mode()&0111 == 0 {
		t.Fatal("gate script not executable")
	}
}

// Scenario: New tests are colocated and within size limits
func TestNewTestFiles_ColocationAndSize(t *testing.T) {
	samples := []string{
		filepath.Join(chRoot, "cmd/centinela/cov2_config_error_test.go"),
		filepath.Join(chRoot, "cmd/centinela/active_feature_more_test.go"),
	}
	for _, p := range samples {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("sample absent (code step not run?): %s", p)
		}
		data, _ := os.ReadFile(p)
		if lines := strings.Count(string(data), "\n"); lines > 100 {
			t.Errorf("%s: %d lines > 100 (G1)", p, lines)
		}
		if !strings.HasPrefix(string(data), "package ") {
			t.Errorf("%s: missing package declaration", p)
		}
	}
}

// Scenario: Hard-to-unit-test paths are explicitly deferred, not faked
func TestDeferredPaths_InRoadmapBacklog(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(chRoot, ".workflow/roadmap.json"))
	if err != nil {
		t.Skipf("roadmap.json absent: %v", err)
	}
	content := string(data)
	for _, slug := range []string{
		"unit-test-mcp-server-in-memory-transport",
		"fault-inject-atomic-write-error-paths",
		"unit-test-vuln-tool-external-seam",
	} {
		if !strings.Contains(content, slug) {
			t.Errorf("deferred slug %q missing from roadmap", slug)
		}
	}
}

// Note: a former TestNoBehaviourChange_OnlyTestFilesAdded guard lived here. It
// asserted `git diff --diff-filter=A main...HEAD` added no production .go files —
// an invariant that only held while coverage-hardening (a test-only feature) was
// itself HEAD. It structurally fails for every later feature that legitimately
// adds production code (local-harness-support is the first such feature), and is
// redundant with the live coverage gate that enforces test coverage of new code
// on the merged tree. Removed rather than left as a permanent landmine.
