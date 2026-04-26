package orchestration

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestValidateEvidencePlanRequiresFeatureSnapshot(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                           //nolint:errcheck
	os.Chdir(d)                                                 //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                              //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                          //nolint:errcheck
	os.MkdirAll("docs/plans", 0755)                             //nolint:errcheck
	os.MkdirAll("specs", 0755)                                  //nolint:errcheck
	os.WriteFile("docs/features/a.md", []byte("a"), 0644)       //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("f"), 0644)       //nolint:errcheck
	os.WriteFile("docs/plans/f.md", []byte("plan"), 0644)       //nolint:errcheck
	os.WriteFile("specs/f.feature", []byte("Feature: f"), 0644) //nolint:errcheck
	path := JSONPath("f", RoleBigThinker)
	bad := `{"feature":"f","step":"plan","role":"big-thinker","status":"done","generatedAt":"` + time.Now().UTC().Format(time.RFC3339) + `","inputs":["docs/features/f.md"],"outputs":["docs/plans/f.md"],"edgeCases":[],"handoffTo":"orchestrator"}`
	os.WriteFile(path, []byte(bad), 0644) //nolint:errcheck
	err := ValidateEvidence(path, "f", "plan", RoleBigThinker, nil)
	if err == nil || !strings.Contains(err.Error(), "missing feature-doc snapshot inputs") {
		t.Fatalf("expected snapshot missing error, got %v", err)
	}
	good := `{"feature":"f","step":"plan","role":"big-thinker","status":"done","generatedAt":"` + time.Now().UTC().Format(time.RFC3339) + `","inputs":["docs/features/f.md","./docs/features/a.md"],"outputs":["docs/plans/f.md"],"edgeCases":[],"handoffTo":"orchestrator"}`
	os.WriteFile(path, []byte(good), 0644) //nolint:errcheck
	if err := ValidateEvidence(path, "f", "plan", RoleBigThinker, nil); err != nil {
		t.Fatalf("expected snapshot success, got %v", err)
	}
	sp := JSONPath("f", RoleFeatureSpecial)
	spec := `{"feature":"f","step":"plan","role":"feature-specialist","status":"done","generatedAt":"` + time.Now().UTC().Format(time.RFC3339) + `","inputs":["docs/features/a.md","docs/features/f.md"],"outputs":["specs/f.feature"],"edgeCases":["e"],"handoffTo":"orchestrator"}`
	os.WriteFile(sp, []byte(spec), 0644) //nolint:errcheck
	if err := ValidateEvidence(sp, "f", "plan", RoleFeatureSpecial, nil); err != nil {
		t.Fatalf("expected feature specialist snapshot success, got %v", err)
	}
}
