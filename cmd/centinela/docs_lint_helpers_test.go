package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/gitdiff"
)

// fakeChangedFiles builds a gitdiff.Resolver whose git calls report exactly
// the given paths as changed, so a CLI test never touches a real repository.
func fakeChangedFiles(paths []string) *gitdiff.Resolver {
	return &gitdiff.Resolver{Run: func(_ string, args ...string) (string, error) {
		if len(args) == 0 {
			return "", nil
		}
		switch args[0] {
		case "merge-base":
			return "0000000000000000000000000000000000000000\n", nil
		case "diff":
			return strings.Join(paths, "\n") + "\n", nil
		}
		return "", nil
	}}
}

// docsLintRepo builds a temp repo whose centinela.toml enables the gate at the
// given severity, and resets the command flags around the test.
func docsLintRepo(t *testing.T, severity string) {
	t.Helper()
	chdir(t, t.TempDir())
	toml := "[gates.docstring]\nenabled = true\nseverity = \"" + severity +
		"\"\nroots = [\"src\"]\n"
	if err := os.WriteFile("centinela.toml", []byte(toml), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("src", 0o755); err != nil {
		t.Fatal(err)
	}
	resetDocsLintFlags()
	t.Cleanup(resetDocsLintFlags)
}

// resetDocsLintFlags clears the package-level cobra flag variables.
func resetDocsLintFlags() {
	docsLintChanged, docsLintFull, docsLintJSON = false, false, false
}

// srcFile writes a Go file under src/ and returns its slash path.
func srcFile(t *testing.T, name, body string) string {
	t.Helper()
	if err := os.WriteFile("src/"+name, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return "src/" + name
}

// stubDocsLintDiff makes resolveDiffFilter report the given changed paths.
func stubDocsLintDiff(t *testing.T, paths ...string) {
	t.Helper()
	prev := gitDiffResolver
	gitDiffResolver = fakeChangedFiles(paths)
	t.Cleanup(func() { gitDiffResolver = prev })
}
