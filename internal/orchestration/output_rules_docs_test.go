package orchestration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The documentation-specialist exemption is removed: outputs must include a
// real file under docs/ (after normalization) or exactly README.md.
func TestDocsSpecialistOutputRuleTable(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, dir := range []string{"docs/guides", "docs/plans"} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range []string{"README.md", "docs/guides/g.md", "docs/plans/f.md", "secret.md"} {
		if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	abs, _ := filepath.Abs("docs/guides/g.md")
	cases := []struct {
		name    string
		outputs []string
		wantErr string // "" = pass
	}{
		{"exemption removed: empty outputs fail", []string{}, "docs/ or README.md"},
		{"docs/ file passes", []string{"docs/guides/g.md"}, ""},
		{"exact README.md passes", []string{"README.md"}, ""},
		{"./README.md normalizes and passes", []string{"./README.md"}, ""},
		{"absolute path fails the prefix rule", []string{abs}, "docs/ or README.md"},
		{"directory is not a file", []string{"docs/guides"}, "real files"},
		{"docs/../secret.md cleans outside docs/", []string{"docs/../secret.md"}, "docs/ or README.md"},
		// Plan decision 3: any existing file under docs/ passes — including the
		// feature's own plan file. Accepted gaming vector, mitigated by prompt
		// duties, pinned here on purpose.
		{"plan-file-only passes (decision 3)", []string{"docs/plans/f.md"}, ""},
	}
	for _, c := range cases {
		err := validateActionableOutputs("x", "f", RoleDocsSpecialist, c.outputs, nil)
		if c.wantErr == "" {
			if err != nil {
				t.Fatalf("%s: want pass, got %v", c.name, err)
			}
			continue
		}
		if err == nil || !strings.Contains(err.Error(), c.wantErr) {
			t.Fatalf("%s: want error containing %q, got %v", c.name, c.wantErr, err)
		}
	}
}
