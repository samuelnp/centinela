package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/centinela-doctor.feature

// Scenario: Doctor runs from inside a worktree and resolves the repo root correctly
func TestDoctorResolvesRepoRootFromWorktree(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	seedHooks(t, dir)
	seedRoadmap(t, dir, "Phase 1: Core")
	wt := filepath.Join(dir, ".worktrees", "some-feature")
	if err := os.MkdirAll(wt, 0o755); err != nil {
		t.Fatal(err)
	}
	out, _ := runDoctor(t, wt)
	// hooks + roadmap are read from the repo root, not the worktree subtree.
	if !strings.Contains(out, "✓ hooks") || !strings.Contains(out, "✓ roadmap") {
		t.Fatalf("checks must operate on canonical root from a worktree:\n%s", out)
	}
}

// Scenario: Doctor runs from repo root and all checks locate their targets
func TestDoctorRunsFromRepoRoot(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	out, _ := runDoctor(t, dir)
	if strings.Contains(out, "missing-file") {
		t.Fatalf("no check should fail on path resolution:\n%s", out)
	}
}

// Scenario: Not inside a git repo causes git-dependent checks to degrade gracefully
func TestDoctorNotInGitRepoDegrades(t *testing.T) {
	dir := doctorRepo(t) // no gitInit
	out, code := runDoctor(t, dir)
	if !strings.Contains(out, "⚠ worktrees") {
		t.Fatalf("worktrees must WARN with no git context:\n%s", out)
	}
	if !strings.Contains(out, "version") {
		t.Fatalf("version check must still produce a diagnosis:\n%s", out)
	}
	// non-git, no errors injected => exit 0, no panic.
	if code != 0 {
		t.Fatalf("degraded non-git run must not error, got %d\n%s", code, out)
	}
}

// Scenario: Doctor does not require an active centinela workflow to run
func TestDoctorRunsWithoutActiveWorkflow(t *testing.T) {
	dir := doctorRepo(t) // empty .workflow, no active workflow
	gitInit(t, dir)
	out, code := runDoctor(t, dir)
	if code != 0 {
		t.Fatalf("no active workflow must still complete cleanly, got %d\n%s", code, out)
	}
	if !strings.Contains(out, "workflow-state") {
		t.Fatalf("checks must still run with no workflow:\n%s", out)
	}
}

// Scenario: Doctor completes without crashing when a check's dependency is missing
func TestDoctorConfigParseErrorDegradesToError(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	writeFile(t, dir, "centinela.toml", "[bad\n") // syntax error
	out, code := runDoctor(t, dir)
	if code != 1 || !strings.Contains(out, "✗ config") {
		t.Fatalf("parse error must surface as config ERROR, got %d\n%s", code, out)
	}
	// other checks still produce diagnoses.
	if !strings.Contains(out, "hooks") || !strings.Contains(out, "evidence") {
		t.Fatalf("non-config checks must still run:\n%s", out)
	}
}
