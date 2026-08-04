// Acceptance: specs/unified-plan-specialist.feature
package acceptance_test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Scenario: A fresh workflow's plan directive names exactly one planner role
func TestUPS_FreshStartPinsPlannerAndDirectiveNamesOnlyPlanner(t *testing.T) {
	bin := upsBuildBin(t)
	dir := upsExistingRepo(t)

	// The orchestration DIRECTIVE is strict-only process (it exists to demand the
	// evidence bundle), so this scenario pins strict explicitly rather than
	// relying on the shipped default, which is now guided.
	mustWrite(t, filepath.Join(dir, "centinela.toml"),
		"[workflow]\nenforcement_profile = \"strict\"\n")

	out, code := runCent(t, bin, dir, "start", "new-widget")
	if code != 0 {
		t.Fatalf("start failed (%d): %s", code, out)
	}
	state := readFile(t, filepath.Join(dir, ".workflow", "new-widget.json"))
	if !strings.Contains(state, `"planContract": "planner-v1"`) {
		t.Fatalf("start must pin planContract=planner-v1: %s", state)
	}

	cmd := exec.Command(bin, "hook", "orchestration")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader("{}")
	directiveOut, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hook orchestration failed: %v\n%s", err, directiveOut)
	}
	directive := string(directiveOut)
	if !strings.Contains(directive, "planner") {
		t.Fatalf("directive must name planner: %s", directive)
	}
	if !strings.Contains(directive, "reasoning") {
		t.Fatalf("directive must carry the reasoning tier: %s", directive)
	}
	if !strings.Contains(directive, ".workflow/new-widget-planner.md") ||
		!strings.Contains(directive, ".workflow/new-widget-planner.json") {
		t.Fatalf("directive must list only the planner evidence pair: %s", directive)
	}
	for _, retired := range []string{"big-thinker", "feature-specialist"} {
		if strings.Contains(directive, retired) {
			t.Fatalf("directive must not mention %q: %s", retired, directive)
		}
	}
}
