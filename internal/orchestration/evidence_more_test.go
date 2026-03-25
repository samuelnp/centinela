package orchestration

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestValidateEvidenceErrorBranches(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)              //nolint:errcheck
	os.Chdir(d)                    //nolint:errcheck
	os.MkdirAll(".workflow", 0755) //nolint:errcheck
	if err := ValidateEvidence(JSONPath("f", RoleBigThinker), "f", "plan", RoleBigThinker); err == nil {
		t.Fatal("expected missing json error")
	}
	path := JSONPath("f", RoleBigThinker)
	os.WriteFile(path, []byte("{bad"), 0644) //nolint:errcheck
	if err := ValidateEvidence(path, "f", "plan", RoleBigThinker); err == nil {
		t.Fatal("expected malformed json error")
	}
	b := baseJSON("f", "plan", string(RoleBigThinker))
	b = strings.Replace(b, `"feature":"f"`, `"feature":"x"`, 1)
	os.WriteFile(path, []byte(b), 0644) //nolint:errcheck
	if err := ValidateEvidence(path, "f", "plan", RoleBigThinker); err == nil {
		t.Fatal("expected mismatched fields error")
	}
}

func TestValidateEvidenceIncompleteAndEdgeCases(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)              //nolint:errcheck
	os.Chdir(d)                    //nolint:errcheck
	os.MkdirAll(".workflow", 0755) //nolint:errcheck
	path := JSONPath("f", RoleBigThinker)
	os.WriteFile(path, []byte(`{"feature":"f","step":"plan","role":"big-thinker","status":"pending","generatedAt":"`+time.Now().UTC().Format(time.RFC3339)+`","inputs":[],"outputs":[],"handoffTo":""}`), 0644) //nolint:errcheck
	if err := ValidateEvidence(path, "f", "plan", RoleBigThinker); err == nil || !strings.Contains(err.Error(), "incomplete") {
		t.Fatalf("expected incomplete error, got %v", err)
	}
	q := JSONPath("f", RoleQASeniorEngineer)
	os.WriteFile(q, []byte(baseJSON("f", "tests", string(RoleQASeniorEngineer))), 0644) //nolint:errcheck
	if err := ValidateEvidence(q, "f", "tests", RoleQASeniorEngineer); err == nil || !strings.Contains(err.Error(), "edgeCases") {
		t.Fatalf("expected edgeCases error, got %v", err)
	}
}

func baseJSON(feature, step, role string) string {
	return `{"feature":"` + feature + `","step":"` + step + `","role":"` + role + `","status":"done","generatedAt":"` + time.Now().UTC().Format(time.RFC3339) + `","inputs":["i"],"outputs":["o"],"edgeCases":[],"handoffTo":"orchestrator"}`
}
