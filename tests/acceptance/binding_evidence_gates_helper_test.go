// Acceptance: specs/binding-evidence-gates.feature
//
// End-to-end coverage of the three gates this feature binds. Every fixture is
// a real temp repo with real workflow state on disk, so each derivation is
// read from a contract rather than a stub; the stamp scenarios drive a binary
// built from ./cmd/centinela. No fixture ever contacts a network remote.
package acceptance_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

// begCfg carries the minimum validate.commands the tests step demands.
func begCfg() *config.Config {
	cfg := &config.Config{}
	cfg.Validate.Commands = []string{"go test ./tests/acceptance/..."}
	return cfg
}

// begRepo chdirs into a temp repo carrying a feature brief and saved workflow.
func begRepo(t *testing.T, feature string, userFacing bool) *workflow.Workflow {
	t.Helper()
	t.Chdir(t.TempDir())
	brief := "# " + feature + "\n"
	if userFacing {
		brief += "surface: user-facing\n"
	}
	mustWrite(t, filepath.Join("docs/features", feature+".md"), brief)
	if err := os.MkdirAll(workflow.WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	return workflow.New(feature)
}

// begSave persists wf after any contract-pin adjustment.
func begSave(t *testing.T, wf *workflow.Workflow) {
	t.Helper()
	if err := workflow.Save(wf); err != nil {
		t.Fatal(err)
	}
}

// begOutputs returns the real files a role's evidence must list, per the
// actionable-outputs rule each role carries.
func begOutputs(t *testing.T, feature string, role orchestration.Role) []string {
	t.Helper()
	var out []string
	switch role {
	case orchestration.RoleQASeniorEngineer:
		out = []string{"tests/unit/x_test.go", filepath.Join(workflow.WorkflowDir, feature+"-edge-cases.md")}
	case orchestration.RoleUXUISpecialist:
		out = []string{"internal/ui/panel.go"}
	case orchestration.RoleDocsSpecialist:
		out = []string{"docs/guides/getting-started.md"}
	case orchestration.RoleMergeSteward:
		out = []string{orchestration.MarkdownPath(feature, role)}
	default:
		out = []string{filepath.Join("internal", string(role)+".go")}
	}
	for _, p := range out {
		if _, err := os.Stat(p); err != nil {
			mustWrite(t, p, "package x\n")
		}
	}
	return out
}

// begEvidence writes a role's complete evidence pair, carrying handoffTo. The
// output files are created so the STRUCTURAL check passes — otherwise the
// chain check never runs and the scenario would prove nothing.
func begEvidence(t *testing.T, feature, step string, role orchestration.Role, handoffTo string) {
	t.Helper()
	outputs := ""
	for i, p := range begOutputs(t, feature, role) {
		if i > 0 {
			outputs += ","
		}
		outputs += fmt.Sprintf("%q", p)
	}
	edges := `"e"`
	if role == orchestration.RoleUXUISpecialist {
		edges = `"mobile-first","visual-hierarchy","typography-hierarchy","responsive-layout",` +
			`"loading-state","empty-state","error-state","motion-and-reduced-motion"`
	}
	mustWrite(t, orchestration.MarkdownPath(feature, role), "# "+string(role)+"\n")
	data := fmt.Sprintf(
		`{"feature":%q,"step":%q,"role":%q,"status":"done","generatedAt":%q,"inputs":["docs/features/%s.md"],"outputs":[%s],"edgeCases":[%s],"mobileFirst":true,"handoffTo":%q}`,
		feature, step, role, time.Now().UTC().Format(time.RFC3339), feature, outputs, edges, handoffTo)
	mustWrite(t, orchestration.JSONPath(feature, role), data)
}

// begTestsStepArtifacts seeds what validateTests demands, so a tests-step
// scenario fails or passes on the HANDOFF rule and nothing else.
func begTestsStepArtifacts(t *testing.T, feature string) {
	t.Helper()
	body := "package acc\n\nfunc TestY(t *testing.T) {\n\tt.Log(\"ran\")\n}\n"
	mustWrite(t, "tests/unit/x_test.go", body)
	mustWrite(t, "tests/acceptance/y_test.go", body)
	mustWrite(t, filepath.Join(workflow.WorkflowDir, feature+"-edge-cases.md"), "# edge cases\n")
}

// begGroundedReport writes a gatekeeper report carrying a well-formed
// commands-run record, so a validate-step scenario is gated on the HANDOFF
// rule rather than on grounding.
func begGroundedReport(t *testing.T, feature string) {
	t.Helper()
	mustWrite(t, workflow.GatekeeperReportPath(feature),
		"**Status:** SAFE\n\n```json centinela:verification\n"+
			`{"revision":"abc123","treeDigest":"sha256:d","commands":[{"argv":["centinela","validate"],"exitCode":0}]}`+
			"\n```\n")
}
