package config

import (
	"strings"
	"testing"
)

// Finding 3 regression: scanned paths arrive repo-relative from git, so an
// unnormalized "./internal" matched nothing and an enforcing gate reported
// "no Go files in scope" on a tree it should have failed. A config spelling
// must never disarm a gate.
func TestNormalizeDocstring_CleansRootSpellings(t *testing.T) {
	got := NormalizeDocstring(DocstringConfig{
		Enabled: true,
		Roots:   []string{"./internal", "cmd/", "src/../src", ".//pkg"},
	})
	want := []string{"internal", "cmd", "src", "pkg"}
	if len(got.Roots) != len(want) {
		t.Fatalf("roots = %v, want %v", got.Roots, want)
	}
	for i := range want {
		if got.Roots[i] != want[i] {
			t.Fatalf("roots = %v, want %v", got.Roots, want)
		}
	}
}

func TestValidateDocstring_RejectsAnAbsoluteRoot(t *testing.T) {
	cfg := NormalizeDocstring(DocstringConfig{
		Enabled: true, Roots: []string{"/Users/me/project/internal"}})
	err := validateDocstring(cfg)
	if err == nil {
		t.Fatal("an absolute root can never match a repo-relative path — fail closed")
	}
	if !strings.Contains(err.Error(), "gates.docstring.roots[0]") ||
		!strings.Contains(err.Error(), "repo-relative") {
		t.Fatalf("error = %q", err.Error())
	}
}
