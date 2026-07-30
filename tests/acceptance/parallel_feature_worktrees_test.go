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

// writeWorktreeSpec drops a spec file into an in-flight feature worktree.
func writeWorktreeSpec(t *testing.T, repo, feat, name, body string) {
	t.Helper()
	dir := filepath.Join(repo, ".worktrees", feat, "specs")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func loginSpec(then string) string {
	return "Feature: Login\n  Scenario: clash\n    Given user has account\n" +
		"    When user logs in\n    Then app routes to " + then + "\n"
}

// Acceptance: specs/parallel-feature-worktrees.feature
// Scenario: Spec conflict across in-flight worktrees is detected before merging
func TestParallelWorktrees_SpecConflictDetectedPreMerge(t *testing.T) {
	repo := initSeedRepo(t)
	// Two in-flight worktrees resolve the SAME scenario in the SAME spec file
	// differently — merging one would silently overwrite the other.
	writeWorktreeSpec(t, repo, "zeta", "login.feature", loginSpec("dashboard"))
	writeWorktreeSpec(t, repo, "eta", "login.feature", loginSpec("onboarding"))

	conflicts := worktree.DetectSpecConflicts(repo, "eta")
	if len(conflicts) != 1 {
		t.Fatalf("expected exactly one spec conflict, got %d: %v", len(conflicts), conflicts)
	}
	msg := worktree.FormatSpecConflicts(conflicts)
	for _, want := range []string{"clash", "zeta", "eta"} {
		if !contains(msg, want) {
			t.Fatalf("formatted conflicts should name %q: %q", want, msg)
		}
	}
}

// Acceptance: specs/parallel-feature-worktrees.feature
// Scenario: Superseding and identical specs never block a merge
// Regression for the spec-conflict-false-positives hotfix, where any existing
// worktree made `centinela merge` unusable.
func TestParallelWorktrees_IdenticalAndSupersedingSpecsDoNotBlock(t *testing.T) {
	repo := initSeedRepo(t)
	specs := filepath.Join(repo, "specs")
	if err := os.MkdirAll(specs, 0755); err != nil {
		t.Fatalf("mkdir specs: %v", err)
	}
	// main carries the old outcome; the merging worktree supersedes it.
	if err := os.WriteFile(filepath.Join(specs, "login.feature"),
		[]byte(loginSpec("dashboard")), 0644); err != nil {
		t.Fatalf("write main spec: %v", err)
	}
	writeWorktreeSpec(t, repo, "eta", "login.feature", loginSpec("onboarding"))
	// A bystander worktree carries a byte-identical copy of main's spec.
	writeWorktreeSpec(t, repo, "bystander", "login.feature", loginSpec("dashboard"))
	// A companion scenario sharing a Given inside one file is normal Gherkin.
	companion := "Feature: Archetypes\n" +
		"  Scenario: hotfix archetype resolves to code-tests-validate\n" +
		"    Given a feature started with the hotfix archetype\n" +
		"    Then the step list is code, tests, validate\n" +
		"  Scenario: active archetype is pinned\n" +
		"    Given a feature started with the hotfix archetype\n" +
		"    Then the archetype is recorded in the workflow state\n"
	writeWorktreeSpec(t, repo, "eta", "archetypes.feature", companion)
	writeWorktreeSpec(t, repo, "bystander", "archetypes.feature", companion)

	if got := worktree.DetectSpecConflicts(repo, "eta"); len(got) != 0 {
		t.Fatalf("supersession and identical copies must not block: %v", got)
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
