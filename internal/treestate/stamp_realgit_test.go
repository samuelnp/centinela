package treestate

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// realRepo builds a temp git repo with one commit and a .gitignore, so Stamp
// runs against actual git output rather than a stub.
func realRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "-q"}, {"config", "user.email", "t@t"}, {"config", "user.name", "t"},
	} {
		c := exec.Command("git", args...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	write(t, dir, "src.go", "package x\n")
	write(t, dir, ".gitignore", "build/\n")
	for _, args := range [][]string{{"add", "-A"}, {"commit", "-qm", "init"}} {
		c := exec.Command("git", args...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	return dir
}

func write(t *testing.T, dir, name, body string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func digestOf(t *testing.T, dir string) string {
	t.Helper()
	snap, err := Stamp(dir, NewExecRunner())
	if err != nil {
		t.Fatal(err)
	}
	return snap.Digest
}

// The remediation loop this feature exists to close: a fix that lands as a
// brand-new, still-untracked package must stale the verification.
func TestStampSeesInsideUntrackedDirectories(t *testing.T) {
	dir := realRepo(t)
	write(t, dir, "newpkg/a.go", "package newpkg\n")
	added := digestOf(t, dir)
	write(t, dir, "newpkg/a.go", "package newpkg // fixed\n")
	if edited := digestOf(t, dir); edited == added {
		t.Fatal("an edit inside an untracked directory must move the digest")
	}
	write(t, dir, "newpkg/b.go", "package newpkg\n")
	if grown := digestOf(t, dir); grown == added {
		t.Fatal("a new file inside an untracked directory must move the digest")
	}
}

// Gitignored paths stay outside the digest BY DESIGN: they are disposable
// build output, never a verification input.
func TestStampIgnoresGitignoredPaths(t *testing.T) {
	dir := realRepo(t)
	base := digestOf(t, dir)
	write(t, dir, "build/out.bin", "artifact\n")
	if withBuild := digestOf(t, dir); withBuild != base {
		t.Fatal("a gitignored build output must not stale the verification")
	}
}

// D3a end to end against real git, not a stub.
func TestStampIgnoresWorkflowChurnAgainstRealGit(t *testing.T) {
	dir := realRepo(t)
	base := digestOf(t, dir)
	write(t, dir, ".workflow/f-gatekeeper.md", "report\n")
	if churned := digestOf(t, dir); churned != base {
		t.Fatal(".workflow/ churn must not stale the verification")
	}
}
