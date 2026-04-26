package orchestration

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestRequiredRolesAndValidateStep(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)              //nolint:errcheck
	os.Chdir(d)                    //nolint:errcheck
	os.MkdirAll(".workflow", 0755) //nolint:errcheck
	if len(RequiredRoles("plan")) != 2 || len(RequiredRoles("code")) != 1 || len(RequiredRoles("docs")) != 1 || len(RequiredRoles("validate")) != 0 {
		t.Fatal("unexpected role mapping")
	}
	if err := ValidateStep("f", "plan"); err == nil {
		t.Fatal("expected missing evidence failure")
	}
	writeEvidence(t, "f", "plan", RoleBigThinker, false)
	writeEvidence(t, "f", "plan", RoleFeatureSpecial, true)
	if err := ValidateStep("f", "plan"); err != nil {
		t.Fatalf("expected valid evidence: %v", err)
	}
	if err := ValidateStep("f", "docs"); err == nil {
		t.Fatal("expected missing docs evidence failure")
	}
	writeEvidence(t, "f", "docs", RoleDocsSpecialist, false)
	if err := ValidateStep("f", "docs"); err != nil {
		t.Fatalf("expected docs evidence success: %v", err)
	}
}

func TestValidateEvidenceBranches(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)              //nolint:errcheck
	os.Chdir(d)                    //nolint:errcheck
	os.MkdirAll(".workflow", 0755) //nolint:errcheck
	path := JSONPath("f", RoleQASeniorEngineer)
	os.WriteFile(path, []byte(`{"feature":"f","step":"tests","role":"qa-senior","status":"done","generatedAt":"bad","inputs":["i"],"outputs":["o"],"edgeCases":["e"],"handoffTo":"orchestrator"}`), 0644) //nolint:errcheck
	if err := ValidateEvidence(path, "f", "tests", RoleQASeniorEngineer); err == nil || !strings.Contains(err.Error(), "generatedAt") {
		t.Fatalf("expected generatedAt error, got %v", err)
	}
}

func writeEvidence(t *testing.T, f, s string, r Role, edge bool) {
	t.Helper()
	os.WriteFile(MarkdownPath(f, r), []byte("# evidence"), 0644) //nolint:errcheck
	edgeCases := `[]`
	inputs := `"inputs":["i"]`
	outputs := `"outputs":["docs/project-docs/index.html"]`
	if s == "plan" && (r == RoleBigThinker || r == RoleFeatureSpecial) {
		inputs = `"inputs":["docs/features/` + f + `.md"]`
		os.MkdirAll("docs/plans", 0755)                              //nolint:errcheck
		os.MkdirAll("specs", 0755)                                   //nolint:errcheck
		os.WriteFile("docs/plans/"+f+".md", []byte("plan"), 0644)    //nolint:errcheck
		os.WriteFile("specs/"+f+".feature", []byte("Feature"), 0644) //nolint:errcheck
		outputs = `"outputs":["docs/plans/` + f + `.md"]`
		if r == RoleFeatureSpecial {
			outputs = `"outputs":["specs/` + f + `.feature"]`
		}
	}
	if s == "docs" {
		os.MkdirAll("docs/project-docs", 0755)                             //nolint:errcheck
		os.WriteFile("docs/project-docs/index.html", []byte("html"), 0644) //nolint:errcheck
	}
	if edge {
		edgeCases = `["e"]`
	}
	data := `{"feature":"` + f + `","step":"` + s + `","role":"` + string(r) + `","status":"done","generatedAt":"` + time.Now().UTC().Format(time.RFC3339) + `",` + inputs + `,` + outputs + `,"edgeCases":` + edgeCases + `,"handoffTo":"orchestrator","extra":"ok"}`
	os.WriteFile(JSONPath(f, r), []byte(data), 0644) //nolint:errcheck
}
