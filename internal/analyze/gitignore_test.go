package analyze

import (
	"path/filepath"
	"testing"
)

func TestLoadGitignore_AbsentIsEmpty(t *testing.T) {
	g := loadGitignore(t.TempDir())
	if g.match("anything") {
		t.Fatal("absent .gitignore must match nothing")
	}
}

func TestGitignore_MatchesPathNameAndDirPrefix(t *testing.T) {
	root := t.TempDir()
	mkFile(t, filepath.Join(root, ".gitignore"),
		"# comment\n!negated\n/leading\nbuildout/\ncoverage.out\n")
	g := loadGitignore(root)
	cases := map[string]bool{
		"leading":            true,  // leading-slash stripped, matches rel
		"buildout":           true,  // dir form, basename match
		"buildout/inner.go":  true,  // dir-prefix hides children
		"coverage.out":       true,  // basename anywhere
		"src/coverage.out":   true,  // basename match deep
		"negated":            false, // negation lines ignored (not added)
		"unrelated/file.go":  false,
	}
	for path, want := range cases {
		if got := g.match(path); got != want {
			t.Fatalf("match(%q) = %v, want %v", path, got, want)
		}
	}
}
