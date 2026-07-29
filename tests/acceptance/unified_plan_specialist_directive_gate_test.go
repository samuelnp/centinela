// Acceptance: specs/unified-plan-specialist.feature
package acceptance_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

// upsDirective runs `hook orchestration` in dir and returns its output.
func upsDirective(t *testing.T, bin, dir string) string {
	t.Helper()
	cmd := exec.Command(bin, "hook", "orchestration")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader("{}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hook orchestration failed: %v\n%s", err, out)
	}
	return string(out)
}

// Scenario: The hook directive and the complete gate name the same required set
func TestUPS_DirectiveAndGateAgree(t *testing.T) {
	bin := upsBuildBin(t)
	cases := []struct {
		feature, contract string
		want              []string
	}{
		{"pinned-workflow", "planner-v1", []string{"planner"}},
		{"unpinned-workflow", "", []string{"big-thinker", "feature-specialist"}},
	}
	for _, tc := range cases {
		dir := upsExistingRepo(t)
		upsWriteWorkflow(t, dir, tc.feature, tc.contract)
		upsFeatureDocs(t, dir, tc.feature)

		directive := upsDirective(t, bin, dir)
		gateOut, code := runCent(t, bin, dir, "complete", tc.feature)
		if code == 0 {
			t.Fatalf("%s: complete with no evidence must fail: %s", tc.feature, gateOut)
		}
		for _, role := range tc.want {
			if !strings.Contains(directive, role) {
				t.Fatalf("%s: directive must name %q: %s", tc.feature, role, directive)
			}
			if !strings.Contains(gateOut, role) {
				t.Fatalf("%s: gate must name %q: %s", tc.feature, role, gateOut)
			}
		}
	}
}

// Scenario: A guided profile with no subagent evidence still names one planner
func TestUPS_GuidedProfileStillNamesPlannerNoEvidencePrinted(t *testing.T) {
	bin := upsBuildBin(t)
	dir := upsExistingRepo(t)
	// OrchestrationMode left empty ("") == guided/outcome: RequiredEvidenceRoles
	// is the contract-aware resolver both the gate and the (skipped) directive
	// share, so it must still resolve to [planner] even though the directive
	// prints nothing for a non-strict workflow.
	body := `{"feature":"guided-feature","currentStep":"plan",` +
		`"stepOrder":["plan","code","tests","validate","docs"],"steps":{},` +
		`"planContract":"planner-v1"}`
	mustWrite(t, dir+"/.workflow/guided-feature.json", body)

	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	roles := workflow.RequiredEvidenceRoles("guided-feature", "plan")
	if len(roles) != 1 || roles[0] != orchestration.RolePlanner {
		t.Fatalf("guided profile must still resolve the single planner role, got %v", roles)
	}
	os.Chdir(orig) //nolint:errcheck

	directive := upsDirective(t, bin, dir)
	if strings.Contains(directive, "Required evidence") {
		t.Fatalf("a non-strict (guided) workflow must print no evidence requirement: %s", directive)
	}
}
