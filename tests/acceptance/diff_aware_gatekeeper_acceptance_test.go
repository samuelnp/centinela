package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/diff-aware-gatekeeper.feature
//
// Each subtest is a Gherkin scenario translated to a binary-level
// invocation against a tmp git repo with a "main" baseline.

func TestDiffAwareGatekeeper(t *testing.T) {
	bin := buildCentinelaBinary(t)

	t.Run("local default is diff-aware", func(t *testing.T) {
		dir := setupGitRepoWithCleanBranch(t)
		out := runValidate(t, bin, dir, nil)
		mustContain(t, out, "Built-in Gates (diff-aware:")
		mustContain(t, out, "since main")
	})

	t.Run("CI default is full scan", func(t *testing.T) {
		dir := setupGitRepoWithCleanBranch(t)
		out := runValidate(t, bin, dir, []string{"CI=true"})
		mustContain(t, out, "Built-in Gates (full scan)")
	})

	t.Run("branch-introduced violation is flagged in diff-aware", func(t *testing.T) {
		dir := setupGitRepoWithCleanBranch(t)
		writeOversized(t, dir, "src/new.go")
		commit(t, dir, "add oversized")
		out := runValidateExpectFail(t, bin, dir, []string{"--changed"})
		mustContain(t, out, "src/new.go")
		mustNotContain(t, out, "src/legacy.go")
	})

	t.Run("untracked oversized file is included", func(t *testing.T) {
		dir := setupGitRepoWithCleanBranch(t)
		writeOversized(t, dir, "src/untracked.go")
		out := runValidateExpectFail(t, bin, dir, []string{"--changed"})
		mustContain(t, out, "src/untracked.go")
	})

	t.Run("full mode reports historical violations", func(t *testing.T) {
		dir := setupGitRepoWithHistoricalViolation(t)
		out := runValidateExpectFail(t, bin, dir, []string{"--full"})
		mustContain(t, out, "Built-in Gates (full scan)")
		mustContain(t, out, "src/legacy.go")
	})

	t.Run("non-git directory degrades to full", func(t *testing.T) {
		dir := t.TempDir()
		out := runValidate(t, bin, dir, []string{"--changed"})
		mustContain(t, out, "diff-aware degraded to full scan")
		mustContain(t, out, "Built-in Gates (full scan)")
	})

	t.Run("mutually exclusive flags rejected", func(t *testing.T) {
		dir := t.TempDir()
		out := runValidateExpectFail(t, bin, dir, []string{"--changed", "--full"})
		mustContain(t, out, "--changed and --full are mutually exclusive")
	})

	t.Run("missing diff base degrades to full with notice", func(t *testing.T) {
		dir := setupGitRepoOnDetachedHEAD(t)
		out := runValidate(t, bin, dir, []string{"--changed"})
		mustContain(t, out, "diff-aware degraded to full scan")
	})

	t.Run("user validate commands always run in full", func(t *testing.T) {
		dir := setupGitRepoWithCleanBranch(t)
		mustWrite(t, filepath.Join(dir, "centinela.toml"),
			"[validate]\ncommands = [\"echo HELLO\"]\ndiff_mode = \"always\"\n")
		out := runValidate(t, bin, dir, []string{"--changed"})
		mustContain(t, out, "HELLO")
		mustContain(t, out, "Built-in Gates (diff-aware:")
	})

	t.Run("configurable diff base honored", func(t *testing.T) {
		dir := setupGitRepoWithCleanBranch(t)
		runGit(t, dir, "branch", "-m", "main", "master")
		mustWrite(t, filepath.Join(dir, "centinela.toml"),
			"[validate]\ndiff_base = \"master\"\n")
		out := runValidate(t, bin, dir, nil)
		mustContain(t, out, "since master")
	})
}

func buildCentinelaBinary(t *testing.T) string {
	t.Helper()
	o, _ := os.Getwd()
	repo := filepath.Clean(filepath.Join(o, "..", ".."))
	bin := filepath.Join(t.TempDir(), "centinela-test")
	build := exec.Command("go", "build", "-o", bin, "./cmd/centinela")
	build.Dir = repo
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build centinela failed: %v\n%s", err, out)
	}
	return bin
}

func setupGitRepoWithCleanBranch(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "config", "user.email", "test@centinela.dev")
	runGit(t, dir, "config", "user.name", "Test")
	mustWrite(t, filepath.Join(dir, "src", "legacy.go"), "package x\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "baseline")
	runGit(t, dir, "checkout", "-q", "-b", "feature")
	return dir
}

func setupGitRepoWithHistoricalViolation(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "config", "user.email", "test@centinela.dev")
	runGit(t, dir, "config", "user.name", "Test")
	mustWrite(t, filepath.Join(dir, "src", "legacy.go"), bigSource(101))
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "historical")
	runGit(t, dir, "checkout", "-q", "-b", "feature")
	return dir
}

func setupGitRepoOnDetachedHEAD(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "other")
	runGit(t, dir, "config", "user.email", "test@centinela.dev")
	runGit(t, dir, "config", "user.name", "Test")
	mustWrite(t, filepath.Join(dir, "src", "x.go"), "package x\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "init")
	return dir
}

func writeOversized(t *testing.T, dir, rel string) {
	t.Helper()
	mustWrite(t, filepath.Join(dir, rel), bigSource(101))
}

func bigSource(lines int) string {
	out := ""
	for i := 0; i < lines; i++ {
		out += "x\n"
	}
	return out
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}

func commit(t *testing.T, dir, msg string) {
	t.Helper()
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", msg)
}

func runValidate(t *testing.T, bin, dir string, env []string) string {
	t.Helper()
	out, err := runValidateRaw(bin, dir, env, nil)
	if err != nil {
		t.Fatalf("validate failed: %v\n%s", err, out)
	}
	return out
}

func runValidateExpectFail(t *testing.T, bin, dir string, extraArgs []string) string {
	t.Helper()
	out, err := runValidateRaw(bin, dir, nil, extraArgs)
	if err == nil {
		t.Fatalf("expected validate to fail, got success:\n%s", out)
	}
	return out
}

func runValidateRaw(bin, dir string, env, extraArgs []string) (string, error) {
	args := []string{"validate"}
	for _, a := range extraArgs {
		if !strings.HasPrefix(a, "--") {
			continue
		}
		args = append(args, a)
	}
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Env = append(append([]string{}, os.Environ()...), "CI=")
	cmd.Env = append(cmd.Env, env...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func mustContain(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Fatalf("expected output to contain %q, got:\n%s", needle, haystack)
	}
}

func mustNotContain(t *testing.T, haystack, needle string) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		t.Fatalf("expected output NOT to contain %q, got:\n%s", needle, haystack)
	}
}
