package acceptance_test

import (
	"os/exec"
	"strings"
	"testing"
)

// mtdaGit runs git in dir, failing the test on error, and returns stdout.
func mtdaGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	c := exec.Command("git", args...)
	c.Dir = dir
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// mtdaSelfMergeRepo builds a repo whose PRIMARY tree has the feature branch
// itself checked out — the shape where a merge would self-no-op and a naive
// ancestor check would fabricate an already-merged success.
func mtdaSelfMergeRepo(t *testing.T, feature string) string {
	t.Helper()
	d := t.TempDir()
	mtdaGit(t, d, "init", "-q", "-b", "main")
	mtdaGit(t, d, "config", "user.email", "qa@centinela.dev")
	mtdaGit(t, d, "config", "user.name", "QA")
	writeFile(t, d, "README.md", "seed\n")
	mtdaGit(t, d, "add", ".")
	mtdaGit(t, d, "commit", "-q", "-m", "seed")
	mtdaGit(t, d, "checkout", "-q", "-b", feature)
	return d
}
