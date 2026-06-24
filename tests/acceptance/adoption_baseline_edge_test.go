// Acceptance: specs/adoption-baseline.feature
package acceptance_test

import (
	"strings"
	"testing"
)

// Scenario: Re-running adopt when a baseline already exists is refused and leaves the file byte-unchanged
func TestAccAdoptSkipIfExists(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true, files: []string{"internal/a.go"}})
	if _, code := runAdopt(t, dir); code != 0 {
		t.Fatal("first adopt should exit 0")
	}
	before := baselineFile(t, dir, "")
	out, code := runAdopt(t, dir)
	if code == 0 {
		t.Fatalf("re-adopt should exit non-zero\n%s", out)
	}
	if !strings.Contains(out, "baseline already exists") || !strings.Contains(out, "--force") {
		t.Fatalf("missing skip guidance: %q", out)
	}
	if before != baselineFile(t, dir, "") {
		t.Fatal("skip must leave the baseline byte-unchanged")
	}
}

// Scenario: Re-running adopt with --force overwrites the existing baseline and exits 0
func TestAccAdoptForce(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true, files: []string{"internal/a.go"}})
	if _, code := runAdopt(t, dir); code != 0 {
		t.Fatal("first adopt should exit 0")
	}
	writeFile(t, dir, "internal/c.go", auditOversized(4))
	out, code := runAdopt(t, dir, "--force")
	if code != 0 {
		t.Fatalf("force adopt exit = %d\n%s", code, out)
	}
	if !strings.Contains(baselineFile(t, dir, ""), "internal/c.go") {
		t.Fatal("force baseline should capture the new violation")
	}
}

// Scenario: Adopting on a clean repo with no violations writes a zero-finding baseline and reports zero accepted findings
func TestAccAdoptCleanRepo(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true})
	out, code := runAdopt(t, dir)
	if code != 0 {
		t.Fatalf("clean adopt exit = %d\n%s", code, out)
	}
	if !strings.Contains(out, "0 accepted findings") || !strings.Contains(out, "nothing to ratchet") {
		t.Fatalf("clean repo report unexpected: %q", out)
	}
}

// Scenario: Adopt --json when a baseline already exists emits a skipped verdict and exits non-zero
func TestAccAdoptJSONSkip(t *testing.T) {
	dir := buildAuditRepo(t, auditRepoBuilder{fileSize: true, files: []string{"internal/a.go"}})
	if _, code := runAdopt(t, dir); code != 0 {
		t.Fatal("first adopt should exit 0")
	}
	before := baselineFile(t, dir, "")
	out, code := runAdopt(t, dir, "--json")
	if code == 0 {
		t.Fatalf("json re-adopt should exit non-zero\n%s", out)
	}
	if !strings.Contains(out, "\"adopted\": false") || !strings.Contains(out, "\"skipped\": true") {
		t.Fatalf("json skip verdict unexpected: %s", out)
	}
	if before != baselineFile(t, dir, "") {
		t.Fatal("json skip must leave the baseline byte-unchanged")
	}
}
