package main

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/workflow"
)

// mcpRepo chdirs into a temp repo carrying a plan-step workflow for feature.
func mcpRepo(t *testing.T, feature, planContract string) {
	t.Helper()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{workflow.WorkflowDir, "docs/features", "docs/plans", "specs"} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	mcpWrite(t, filepath.Join("docs/features", feature+".md"), "# brief\n")
	mcpWrite(t, filepath.Join("docs/plans", feature+".md"), "# plan\n")
	wf := workflow.New(feature)
	wf.PlanContract = planContract
	if err := workflow.Save(wf); err != nil {
		t.Fatal(err)
	}
}

func mcpWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// mcpEvidence writes a readable evidence pair for role at step.
func mcpEvidence(t *testing.T, feature, step, role string) {
	t.Helper()
	ts := time.Now().UTC().Format(time.RFC3339)
	mcpWrite(t, filepath.Join(workflow.WorkflowDir, feature+"-"+role+".md"), "# "+role+"\n")
	mcpWrite(t, filepath.Join(workflow.WorkflowDir, feature+"-"+role+".json"),
		`{"feature":"`+feature+`","step":"`+step+`","role":"`+role+`","status":"done",`+
			`"generatedAt":"`+ts+`","inputs":["docs/features/`+feature+`.md"],`+
			`"outputs":["docs/plans/`+feature+`.md"],"edgeCases":["e"],"mobileFirst":true,`+
			`"handoffTo":"senior-engineer"}`)
}

// mcpVerifiedRoles returns the sorted role set the MCP verdict actually checked.
// It FAILS on the "no claims to verify" placeholder: that placeholder is exactly
// what a contract-blind lookup produces, and asserting only role ABSENCE over it
// is vacuous because the loop body never runs.
func mcpVerifiedRoles(t *testing.T, feature string) []string {
	t.Helper()
	pkt, err := mcpVerdict(feature)
	if err != nil {
		t.Fatalf("mcpVerdict: %v", err)
	}
	if len(pkt.Verify) == 0 {
		t.Fatal("MCP verdict verified nothing")
	}
	seen := map[string]bool{}
	var roles []string
	for _, c := range pkt.Verify {
		if c.Claim == "claims" {
			t.Fatalf("MCP verdict reported the no-claims placeholder: %+v", c)
		}
		if c.Role != "" && !seen[c.Role] {
			seen[c.Role] = true
			roles = append(roles, c.Role)
		}
	}
	sort.Strings(roles)
	return roles
}
