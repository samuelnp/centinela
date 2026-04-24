package docgen

import (
	"os"
	"strings"
	"testing"
)

func TestGenerateAndRender(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	writeFixture(t)
	if err := Generate("docs/project-docs/index.html", "Doc"); err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	out, err := os.ReadFile("docs/project-docs/index.html")
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	s := string(out)
	for _, want := range []string{"Project Overview", "Latest Features", "Getting Started", "Feature Topology", "Documentation Examples", "class=\"mermaid\""} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q", want)
		}
	}
}
