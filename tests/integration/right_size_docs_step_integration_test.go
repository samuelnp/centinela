package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

func rdsIntWrite(t *testing.T, rel, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(rel), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(rel, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// An internal feature reaches a valid docs step with only a one-line changelog:
// no documentation-specialist role and no knowledge-base bundle required.
func TestRDSIntegration_InternalLightPath(t *testing.T) {
	t.Chdir(t.TempDir())
	rdsIntWrite(t, "docs/features/in.md", "# in\n")
	rdsIntWrite(t, workflow.WorkflowDir+"/in-changelog.md", "- refactor: tidy docs step\n")

	if roles := orchestration.RequiredRolesForFeature("in", "docs"); len(roles) != 0 {
		t.Fatalf("internal docs step must require no roles, got %v", roles)
	}
	if err := workflow.ValidateArtifacts("in", "docs", nil); err != nil {
		t.Fatalf("internal docs step must pass with only a changelog, got %v", err)
	}
}

// A user-facing feature still needs the full bundle: documentation-specialist
// role plus portal + knowledge-base. A bare changelog is not enough.
func TestRDSIntegration_UserFacingNeedsFullBundle(t *testing.T) {
	t.Chdir(t.TempDir())
	rdsIntWrite(t, "docs/features/uf.md", "# uf\nsurface: user-facing\n")
	rdsIntWrite(t, workflow.WorkflowDir+"/uf-changelog.md", "- feat: x\n")

	if roles := orchestration.RequiredRolesForFeature("uf", "docs"); len(roles) == 0 {
		t.Fatal("user-facing docs step must still require the documentation-specialist role")
	}
	err := workflow.ValidateArtifacts("uf", "docs", nil)
	if err == nil || !strings.Contains(err.Error(), "documentation output") {
		t.Fatalf("user-facing docs step must demand the portal/KB bundle, got %v", err)
	}
}
