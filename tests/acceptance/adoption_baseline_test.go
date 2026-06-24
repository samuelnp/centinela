// Acceptance: specs/adoption-baseline.feature
package acceptance_test

import (
	"strings"
	"testing"
)

// runAdopt runs `centinela adopt` in dir, reusing the shared cent binary runner.
func runAdopt(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	return runCent(t, buildCent(t), dir, append([]string{"adopt"}, args...)...)
}

// Scenario: First adoption on a repo with pre-existing violations records the baseline and reports the accepted debt
func TestAccAdoptFirstAdoption(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true, files: []string{"internal/a.go", "internal/b.go"}})
	out, code := runAdopt(t, dir)
	if code != 0 {
		t.Fatalf("exit = %d, want 0\n%s", code, out)
	}
	if !strings.Contains(out, "Adopted baseline") || !strings.Contains(out, "ratchet to zero") {
		t.Fatalf("missing adoption report: %q", out)
	}
	if b := baselineFile(t, dir, ""); !strings.Contains(b, "G1: File Size") {
		t.Fatalf("baseline missing recorded gate: %s", b)
	}
}

// Scenario: After adoption a fresh audit reports zero new violations so day-one validate is not drowned
func TestAccAdoptThenAuditClean(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true, files: []string{"internal/a.go"}})
	if _, code := runAdopt(t, dir); code != 0 {
		t.Fatal("adopt should exit 0")
	}
	out, code := runCent(t, buildCent(t), dir, "audit")
	if code != 0 {
		t.Fatalf("post-adoption audit exit = %d\n%s", code, out)
	}
	if !strings.Contains(out, "0 new") {
		t.Fatalf("expected 0 new after adoption: %q", out)
	}
}

// Scenario: Adopt with --json emits a machine-readable adoption summary instead of the human report
func TestAccAdoptJSON(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true, files: []string{"internal/a.go"}})
	out, code := runAdopt(t, dir, "--json")
	if code != 0 {
		t.Fatalf("json adopt exit = %d\n%s", code, out)
	}
	for _, want := range []string{"\"adopted\": true", "\"skipped\": false", "audit-baseline.json", "\"per_gate\""} {
		if !strings.Contains(out, want) {
			t.Fatalf("json verdict missing %q: %s", want, out)
		}
	}
	if strings.Contains(out, "ratchet to zero") {
		t.Fatalf("json mode should not print the human report: %s", out)
	}
}

// Scenario: The baseline written by adopt is byte-identical to audit.Record plus audit.Save output
func TestAccAdoptDeterministic(t *testing.T) {
	a := buildAuditRepo(t, auditRepoBuilder{fileSize: true, files: []string{"internal/a.go"}})
	b := buildAuditRepo(t, auditRepoBuilder{fileSize: true, files: []string{"internal/a.go"}})
	if _, code := runAdopt(t, a); code != 0 {
		t.Fatal("adopt a failed")
	}
	if _, code := runAdopt(t, b); code != 0 {
		t.Fatal("adopt b failed")
	}
	if baselineFile(t, a, "") != baselineFile(t, b, "") {
		t.Fatal("adopt over identical repos must yield byte-identical baselines")
	}
}
