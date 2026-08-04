package workflow

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
)

// A strict user-facing docs step demands documentation-specialist evidence
// whose outputs include a real updated file under docs/ or README.md.
func TestValidateArtifactsDocsStrictOrchestration(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                             //nolint:errcheck
	os.Chdir(d)                                                                   //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                                //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                                            //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("surface: user-facing\n"), 0644)    //nolint:errcheck
	os.WriteFile(".workflow/f-changelog.md", []byte("- feat: real docs\n"), 0644) //nolint:errcheck
	os.MkdirAll("docs/guides", 0755)                                              //nolint:errcheck
	os.WriteFile("docs/guides/getting-started.md", []byte("# Guide\nnew"), 0644)  //nolint:errcheck
	wf := New("f")
	wf.CurrentStep = "docs"
	Save(wf) //nolint:errcheck
	if err := ValidateArtifacts("f", "docs", &config.Config{}); err == nil {
		t.Fatal("expected docs orchestration evidence failure")
	}
	os.WriteFile(orchestration.MarkdownPath("f", orchestration.RoleDocsSpecialist), []byte("# role"), 0644) //nolint:errcheck
	// Outputs listing only .workflow/ evidence fail the real-doc-file rule.
	writeDocsEvidence(t, `".workflow/f-documentation-specialist.md"`)
	err := ValidateArtifacts("f", "docs", &config.Config{})
	if err == nil || !strings.Contains(err.Error(), "docs/ or README.md") {
		t.Fatalf("expected real-doc-file rule failure, got %v", err)
	}
	// A real updated file under docs/ satisfies the rule.
	writeDocsEvidence(t, `"docs/guides/getting-started.md"`)
	if err := ValidateArtifacts("f", "docs", &config.Config{}); err != nil {
		t.Fatalf("expected docs strict success, got %v", err)
	}
}

func writeDocsEvidence(t *testing.T, outputs string) {
	t.Helper()
	data := `{"feature":"f","step":"docs","role":"documentation-specialist","status":"done","generatedAt":"` +
		time.Now().UTC().Format(time.RFC3339) + `","inputs":["i"],"outputs":[` + outputs + `],"edgeCases":[],"handoffTo":"complete"}`
	os.WriteFile(orchestration.JSONPath("f", orchestration.RoleDocsSpecialist), []byte(data), 0644) //nolint:errcheck
}
