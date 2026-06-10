package workflow

import (
	"os"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
)

func TestValidateArtifactsDocsStrictOrchestration(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                          //nolint:errcheck
	os.Chdir(d)                                                                //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                             //nolint:errcheck
	os.MkdirAll("docs/project-docs/kb", 0755)                                  //nolint:errcheck
	os.WriteFile("docs/project-docs/index.html", []byte("ok"), 0644)           //nolint:errcheck
	os.WriteFile("docs/project-docs/kb/f.md", []byte("ok"), 0644)              //nolint:errcheck
	os.WriteFile("docs/project-docs/kb/f.html", []byte("<html></html>"), 0644) //nolint:errcheck
	wf := New("f")
	wf.CurrentStep = "docs"
	Save(wf) //nolint:errcheck
	if err := ValidateArtifacts("f", "docs", &config.Config{}); err == nil {
		t.Fatal("expected docs orchestration evidence failure")
	}
	os.WriteFile(orchestration.MarkdownPath("f", orchestration.RoleDocsSpecialist), []byte("# role"), 0644) //nolint:errcheck
	data := `{"feature":"f","step":"docs","role":"documentation-specialist","status":"done","generatedAt":"` + time.Now().UTC().Format(time.RFC3339) + `","inputs":["i"],"outputs":["o"],"edgeCases":[],"handoffTo":"orchestrator"}`
	os.WriteFile(orchestration.JSONPath("f", orchestration.RoleDocsSpecialist), []byte(data), 0644) //nolint:errcheck
	if err := ValidateArtifacts("f", "docs", &config.Config{}); err != nil {
		t.Fatalf("expected docs strict success, got %v", err)
	}
}
