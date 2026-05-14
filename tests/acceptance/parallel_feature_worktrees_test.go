package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/worktree"
)

// initSeedRepo mirrors the wizard contract: git init main, README + .gitignore
// containing .worktrees/, and one seed commit.
func initSeedRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		c := exec.Command("git", args...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q", "-b", "main")
	run("config", "user.email", "qa@centinela.dev")
	run("config", "user.name", "QA")
	_ = os.WriteFile(filepath.Join(dir, "README.md"), []byte("seed\n"), 0644)
	_ = os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(".worktrees/\n"), 0644)
	run("add", ".")
	run("commit", "-q", "-m", "seed")
	return dir
}

// Acceptance: specs/parallel-feature-worktrees.feature
// Scenario: Start provisions a worktree when use_worktrees is enabled.
func TestParallelWorktrees_StartProvisionsWhenEnabled(t *testing.T) {
	repo := initSeedRepo(t)
	cfg := &config.Config{}
	cfg.Workflow.UseWorktrees = true
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(repo)       //nolint:errcheck

	path, err := worktree.MaybeProvision(repo, "alpha", cfg)
	if err != nil {
		t.Fatalf("MaybeProvision: %v", err)
	}
	if path == "" {
		t.Fatal("expected a worktree path when use_worktrees=true")
	}
	if _, err := os.Stat(filepath.Join(repo, ".worktrees", "alpha")); err != nil {
		t.Fatalf("worktree dir missing on disk: %v", err)
	}
}

// Acceptance: Scenario: Start runs in the main checkout when use_worktrees is disabled.
func TestParallelWorktrees_StartNoProvisionWhenDisabled(t *testing.T) {
	repo := initSeedRepo(t)
	cfg := &config.Config{}
	cfg.Workflow.UseWorktrees = false
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(repo)       //nolint:errcheck

	path, err := worktree.MaybeProvision(repo, "beta", cfg)
	if err != nil {
		t.Fatalf("MaybeProvision should not error when flag is off: %v", err)
	}
	if path != "" {
		t.Fatalf("expected empty path when flag is off, got %q", path)
	}
	if _, err := os.Stat(filepath.Join(repo, ".worktrees")); !os.IsNotExist(err) {
		t.Fatalf(".worktrees/ should not exist when flag is off; stat err=%v", err)
	}
}

// Acceptance: Scenario: Migrate syncs tool ignore lists for existing projects.
func TestParallelWorktrees_MigrateSyncsIgnoreFilesIdempotently(t *testing.T) {
	repo := t.TempDir()
	// Pre-existing project without .worktrees/ entries anywhere.
	_ = os.WriteFile(filepath.Join(repo, ".gitignore"), []byte("node_modules/\n"), 0644)
	_ = os.WriteFile(filepath.Join(repo, ".eslintignore"), []byte("dist/\n"), 0644)

	res, err := worktree.SyncIgnores(repo)
	if err != nil {
		t.Fatalf("first SyncIgnores: %v", err)
	}
	if len(res.Touched) == 0 {
		t.Fatal("first sync should patch at least one file")
	}
	gi, _ := os.ReadFile(filepath.Join(repo, ".gitignore"))
	if !contains(string(gi), ".worktrees/") {
		t.Fatalf(".gitignore should contain .worktrees/: %q", gi)
	}
	// Idempotent: second run is a no-op.
	res2, err := worktree.SyncIgnores(repo)
	if err != nil {
		t.Fatalf("second SyncIgnores: %v", err)
	}
	if len(res2.Touched) != 0 {
		t.Fatalf("second sync must touch nothing, got %v", res2.Touched)
	}
}

// Acceptance: Scenario: Spec conflict across in-flight worktrees is detected
// before merging.
func TestParallelWorktrees_SpecConflictDetectedPreMerge(t *testing.T) {
	repo := initSeedRepo(t)
	// Two worktrees, each with a contradictory scenario keyed off the same Given.
	specsMain := filepath.Join(repo, "specs")
	_ = os.MkdirAll(specsMain, 0755)
	zeta := "Feature: Zeta\n  Scenario: clash\n    Given user has account\n    When user logs in\n    Then app routes to dashboard\n"
	_ = os.WriteFile(filepath.Join(specsMain, "zeta.feature"), []byte(zeta), 0644)

	// Simulate an in-flight 'eta' worktree with a clashing scenario.
	etaWT := filepath.Join(repo, ".worktrees", "eta", "specs")
	_ = os.MkdirAll(etaWT, 0755)
	eta := "Feature: Eta\n  Scenario: clash\n    Given user has account\n    When user logs in\n    Then app routes to onboarding\n"
	_ = os.WriteFile(filepath.Join(etaWT, "eta.feature"), []byte(eta), 0644)

	conflicts := worktree.DetectSpecConflicts(repo, "eta")
	if len(conflicts) == 0 {
		t.Fatal("expected at least one spec conflict, got none")
	}
	msg := worktree.FormatSpecConflicts(conflicts)
	if !contains(msg, "clash") {
		t.Fatalf("formatted conflicts should name the scenario: %q", msg)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
