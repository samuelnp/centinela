// Acceptance: specs/unified-plan-specialist.feature
package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// Scenario: Plan-advisor phrases its questions as two lenses for one agent
func TestUPS_PlanAdvisorTwoLensesOneAgent(t *testing.T) {
	bin := upsBuildBin(t)
	dir := upsExistingRepo(t)
	if err := os.MkdirAll(filepath.Join(dir, "docs", "features"), 0o755); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(dir, "docs", "features", "f.md"), "surface: user-facing\n")
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	if err := workflow.Save(workflow.New("f")); err != nil {
		t.Fatal(err)
	}
	os.Chdir(orig) //nolint:errcheck

	out, code := runCent(t, bin, dir, "hook", "plan-advisor")
	if code != 0 {
		t.Fatalf("plan-advisor hook failed: %s", out)
	}
	if !strings.Contains(out, "One planner agent, two lenses: strategy first, then spec.") {
		t.Fatalf("advisor must render the one-agent, two-lens header: %s", out)
	}
	if !strings.Contains(out, "[strategy]") && !strings.Contains(out, "[spec]") {
		t.Fatalf("advisor must tag questions [strategy] or [spec]: %s", out)
	}
	for _, retired := range []string{"[big-thinker]", "[feature-specialist]"} {
		if strings.Contains(out, retired) {
			t.Fatalf("advisor must not tag %q: %s", retired, out)
		}
	}
	if strings.Contains(out, "delegate to two") || strings.Contains(out, "two separate agents") {
		t.Fatalf("advisor must not instruct delegating to two separate agents: %s", out)
	}
}
