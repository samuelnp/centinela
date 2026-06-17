package acceptance_test

import (
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/centinela-doctor.feature

// Scenario: Orphaned *.json.tmp files are flagged and removed by --fix
func TestDoctorEvidenceTmpFlaggedAndRemoved(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	writeFile(t, dir, ".workflow/feat-qa-senior.json.tmp", "{}")
	out, code := runDoctor(t, dir)
	if code != 1 || !strings.Contains(out, "✗ evidence") {
		t.Fatalf("tmp must Error/exit 1, got %d\n%s", code, out)
	}
	if !strings.Contains(out, ".json.tmp") {
		t.Fatalf("must list the tmp path:\n%s", out)
	}
	out2, code2 := runDoctor(t, dir, "--fix")
	if code2 != 0 || strings.Contains(out2, "✗ evidence") {
		t.Fatalf("--fix must remove tmp files, got %d\n%s", code2, out2)
	}
	left, _ := filepath.Glob(filepath.Join(dir, ".workflow", "*.json.tmp"))
	if len(left) != 0 {
		t.Fatalf("all tmp files must be removed, left %v", left)
	}
}

// Scenario: Orphaned evidence removal under --fix is idempotent
func TestDoctorEvidenceFixIdempotent(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	writeFile(t, dir, ".workflow/feat-qa-senior.json.tmp", "{}")
	runDoctor(t, dir, "--fix")
	out, code := runDoctor(t, dir, "--fix")
	if code != 0 || !strings.Contains(out, "✓ evidence") {
		t.Fatalf("idempotent --fix must keep evidence OK, got %d\n%s", code, out)
	}
}
