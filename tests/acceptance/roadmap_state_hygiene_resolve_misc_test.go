// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Scenario: resolve regenerates when only ROADMAP.md is conflicted
func TestRsh_ResolveRegeneratesAMarkdownOnlyConflict(t *testing.T) {
	bin := buildCent(t)
	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "config", "user.email", "rsh@centinela.dev")
	runGit(t, dir, "config", "user.name", "RSH")
	runGit(t, dir, "config", "commit.gpgsign", "false")
	rm := filepath.Join(dir, ".workflow", "roadmap.json")
	md := filepath.Join(dir, "ROADMAP.md")
	mustWrite(t, rm, rshBaseRoadmap) // identical on both sides: merges cleanly
	mustWrite(t, md, "# Roadmap\n\nbase\n")
	commit(t, dir, "base")

	runGit(t, dir, "checkout", "-q", "-b", "theirs")
	mustWrite(t, md, "# Roadmap\n\ntheirs\n")
	commit(t, dir, "theirs")

	runGit(t, dir, "checkout", "-q", "main")
	mustWrite(t, md, "# Roadmap\n\nours\n")
	commit(t, dir, "ours")
	c := exec.Command("git", "merge", "--no-edit", "theirs")
	c.Dir = dir
	_, _ = c.CombinedOutput() // a conflicting merge exits non-zero — expected

	if rshUnmerged(t, dir) {
		t.Fatal("fixture bug: roadmap.json must merge cleanly here")
	}
	if rshGitOut(t, dir, "ls-files", "--unmerged", "--", "ROADMAP.md") == "" {
		t.Fatal("fixture bug: ROADMAP.md must be the conflicted path")
	}

	out, code := runCent(t, bin, dir, "roadmap", "resolve")
	if code != 0 {
		t.Fatalf("resolve exit=%d\n%s", code, out)
	}
	body := mustRead(t, md)
	if strings.Contains(body, "<<<<<<<") {
		t.Fatalf("no markers may survive:\n%s", body)
	}
	if got := rshGitOut(t, dir, "diff", "--cached", "--name-only"); got != "ROADMAP.md" {
		t.Fatalf("ROADMAP.md must be staged, got %q", got)
	}
}

// Scenario: resolve is a no-op outside a conflict
func TestRsh_ResolveIsANoopOutsideAConflict(t *testing.T) {
	bin := buildCent(t)
	dir := rshRepo(t, rshBaseRoadmap)
	before := mustRead(t, dir+"/.workflow/roadmap.json")

	out, code := runCent(t, bin, dir, "roadmap", "resolve")
	if code != 0 {
		t.Fatalf("exit=%d\n%s", code, out)
	}
	containsAll(t, out, "Nothing to resolve")
	if got := mustRead(t, dir+"/.workflow/roadmap.json"); got != before {
		t.Fatal("roadmap.json must not be modified")
	}
	if got := rshGitOut(t, dir, "status", "--porcelain"); got != "" {
		t.Fatalf("nothing may be modified or staged, got %q", got)
	}
}
