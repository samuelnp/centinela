package unit_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func TestUXEvidenceRequiresMobileFirst(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                   //nolint:errcheck
	os.Chdir(d)                                         //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                      //nolint:errcheck
	os.MkdirAll("src/ui", 0755)                         //nolint:errcheck
	os.WriteFile("src/ui/page.tsx", []byte("ok"), 0644) //nolint:errcheck
	path := orchestration.JSONPath("f", orchestration.RoleUXUISpecialist)
	ts := time.Now().UTC().Format(time.RFC3339)
	data := `{"feature":"f","step":"code","role":"ux-ui-specialist","status":"done","generatedAt":"` + ts + `","inputs":["docs/features/f.md"],"outputs":["src/ui/page.tsx"],"edgeCases":["mobile-first","visual-hierarchy","typography-hierarchy","responsive-layout","loading-state","empty-state","error-state","motion-and-reduced-motion"],"handoffTo":"qa-senior"}`
	os.WriteFile(path, []byte(data), 0644) //nolint:errcheck
	err := orchestration.ValidateEvidence(path, "f", "code", orchestration.RoleUXUISpecialist, []string{"src/ui"})
	if err == nil || !strings.Contains(err.Error(), "mobileFirst") {
		t.Fatalf("expected mobileFirst validation error, got %v", err)
	}
}
