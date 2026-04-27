package integration_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

func TestUserFacingCodeStepRequiresUXTags(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                               //nolint:errcheck
	os.Chdir(d)                                                                     //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755)                                         //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                                              //nolint:errcheck
	os.MkdirAll("src/ui", 0755)                                                     //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("surface: user-facing\n"), 0644)      //nolint:errcheck
	os.WriteFile("src/ui/page.tsx", []byte("export const Page = () => null"), 0644) //nolint:errcheck
	wf := workflow.New("f")
	wf.CurrentStep = "code"
	workflow.Save(wf) //nolint:errcheck
	ts := time.Now().UTC().Format(time.RFC3339)
	os.WriteFile(orchestration.MarkdownPath("f", orchestration.RoleSeniorEngineer), []byte("# role"), 0644) //nolint:errcheck
	os.WriteFile(orchestration.MarkdownPath("f", orchestration.RoleUXUISpecialist), []byte("# role"), 0644) //nolint:errcheck
	se := `{"feature":"f","step":"code","role":"senior-engineer","status":"done","generatedAt":"` + ts + `","inputs":["docs/plans/f.md"],"outputs":["src/ui/page.tsx"],"edgeCases":[],"handoffTo":"qa-senior"}`
	ux := `{"feature":"f","step":"code","role":"ux-ui-specialist","status":"done","generatedAt":"` + ts + `","inputs":["docs/features/f.md"],"outputs":["src/ui/page.tsx"],"edgeCases":["mobile-first"],"mobileFirst":true,"handoffTo":"qa-senior"}`
	os.WriteFile(orchestration.JSONPath("f", orchestration.RoleSeniorEngineer), []byte(se), 0644) //nolint:errcheck
	os.WriteFile(orchestration.JSONPath("f", orchestration.RoleUXUISpecialist), []byte(ux), 0644) //nolint:errcheck
	err := workflow.ValidateArtifacts("f", "code", &config.Config{})
	if err == nil || !strings.Contains(err.Error(), "visual-hierarchy") {
		t.Fatalf("expected missing ux tag error, got %v", err)
	}
}
