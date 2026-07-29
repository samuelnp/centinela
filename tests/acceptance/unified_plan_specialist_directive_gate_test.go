// Acceptance: specs/unified-plan-specialist.feature
package acceptance_test

import (
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"sort"
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

// upsEvidenceRoles extracts the role set an operator-facing surface actually
// demands, by scanning the `.workflow/<feature>-<role>.{md,json}` paths BOTH the
// directive ("Required evidence before ...") and the gate ("missing: ...") print.
// Reading the evidence paths rather than substring-matching role names is what
// makes set EQUALITY assertable: the unpinned gate's contract annotation mentions
// the literal "planner-v1", which a naive !Contains("planner") would trip over.
func upsEvidenceRoles(out, feature string) []string {
	re := regexp.MustCompile(`\.workflow/` + regexp.QuoteMeta(feature) + `-([a-z0-9-]+)\.(?:md|json)`)
	seen := map[string]bool{}
	var roles []string
	for _, m := range re.FindAllStringSubmatch(out, -1) {
		if seen[m[1]] {
			continue
		}
		seen[m[1]] = true
		roles = append(roles, m[1])
	}
	sort.Strings(roles)
	return roles
}

// Scenario: The hook directive and the complete gate name the same required set
//
// (a Scenario Outline in the spec; the traceability marker must read
// "// Scenario:" verbatim — the gate's regex does not match "Scenario Outline:")
//
// The Then clauses say the sets are EQUAL, so this asserts equality in both
// directions — the directive set == the gate set == the expected set — rather
// than mere containment. A containment-only guard would not catch a directive
// that named a SUPERSET, which is exactly the PR #83 divergence class.
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

		want := append([]string(nil), tc.want...)
		sort.Strings(want)
		gotDirective := upsEvidenceRoles(directive, tc.feature)
		gotGate := upsEvidenceRoles(gateOut, tc.feature)

		if !reflect.DeepEqual(gotDirective, want) {
			t.Fatalf("%s: directive role set = %v, want exactly %v\n%s",
				tc.feature, gotDirective, want, directive)
		}
		if !reflect.DeepEqual(gotGate, want) {
			t.Fatalf("%s: gate role set = %v, want exactly %v\n%s",
				tc.feature, gotGate, want, gateOut)
		}
		// The equality that actually matters: the two surfaces to each other.
		if !reflect.DeepEqual(gotDirective, gotGate) {
			t.Fatalf("%s: directive %v and gate %v disagree", tc.feature, gotDirective, gotGate)
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
