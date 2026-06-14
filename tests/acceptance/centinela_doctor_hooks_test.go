package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/centinela-doctor.feature

// Scenario: Healthy project reports all checks OK and exits 0
func TestDoctorHealthyAllOK(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	seedHooks(t, dir)
	out, code := runDoctor(t, dir)
	if code != 0 {
		t.Fatalf("healthy must exit 0, got %d\n%s", code, out)
	}
	if strings.Contains(out, "✗") {
		t.Fatalf("healthy must have no errors:\n%s", out)
	}
	if !strings.Contains(out, "0 error") {
		t.Fatalf("summary must report 0 error:\n%s", out)
	}
}

// Scenario: Missing hook entries are flagged as ERROR and --fix re-wires them
func TestDoctorMissingHooksErrorThenFix(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	writeFile(t, dir, ".claude/settings.json", "{}")
	out, code := runDoctor(t, dir)
	if code != 1 || !strings.Contains(out, "✗ hooks") {
		t.Fatalf("missing hooks must Error/exit 1, got %d\n%s", code, out)
	}
	out2, code2 := runDoctor(t, dir, "--fix")
	if code2 != 0 || strings.Contains(out2, "✗ hooks") {
		t.Fatalf("--fix must re-wire hooks, got %d\n%s", code2, out2)
	}
	data, _ := os.ReadFile(filepath.Join(dir, ".claude/settings.json"))
	if !strings.Contains(string(data), "hooks") {
		t.Fatal("settings.json must contain hook entries after --fix")
	}
}

// Scenario: Hook re-wire under --fix is idempotent
func TestDoctorHookFixIdempotent(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	writeFile(t, dir, ".claude/settings.json", "{}")
	runDoctor(t, dir, "--fix")
	before, _ := os.ReadFile(filepath.Join(dir, ".claude/settings.json"))
	out, code := runDoctor(t, dir, "--fix")
	after, _ := os.ReadFile(filepath.Join(dir, ".claude/settings.json"))
	if string(before) != string(after) {
		t.Fatal("second --fix must leave settings.json byte-identical")
	}
	if code != 0 || strings.Contains(out, "✗ hooks") {
		t.Fatalf("idempotent --fix must stay OK, got %d\n%s", code, out)
	}
}

// Scenario: No .claude directory causes hooks check to degrade to WARN not crash
func TestDoctorNoClaudeDirWarns(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	out, code := runDoctor(t, dir)
	if code != 0 || !strings.Contains(out, "⚠ hooks") {
		t.Fatalf("no .claude must WARN/exit 0, got %d\n%s", code, out)
	}
	if !strings.Contains(out, "centinela setup") {
		t.Fatalf("hooks WARN must mention setup:\n%s", out)
	}
}
