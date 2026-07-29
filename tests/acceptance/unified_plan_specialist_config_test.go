// Acceptance: specs/unified-plan-specialist.feature
package acceptance_test

import (
	"os/exec"
	"strings"
	"testing"
)

const upsLegacyConfigTOML = "[orchestration.models]\nbig-thinker = \"reasoning\"\nfeature-specialist = \"balanced\"\n"

// Scenario: A legacy per-role model override key still resolves, aliased to planner
func TestUPS_LegacyModelKeyAliasesToPlanner(t *testing.T) {
	bin := upsBuildBin(t)
	dir := upsExistingRepo(t)
	mustWrite(t, dir+"/centinela.toml", upsLegacyConfigTOML)
	upsWriteWorkflow(t, dir, "new-widget", "planner-v1")

	cmd := exec.Command(bin, "hook", "orchestration")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader("{}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hook orchestration failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "planner (model: ") || !strings.Contains(string(out), "reasoning") {
		t.Fatalf("the big-thinker override must alias onto planner: %s", out)
	}
}

// Scenario: The legacy-key deprecation notice appears at doctor and start, not on every hook call
func TestUPS_DeprecationNoticeAtDoctorAndStart_NeverOnHook(t *testing.T) {
	bin := upsBuildBin(t)
	dir := upsExistingRepo(t)
	mustWrite(t, dir+"/centinela.toml", upsLegacyConfigTOML)

	doctorOut, _ := runCent(t, bin, dir, "doctor")
	if !strings.Contains(doctorOut, "planner") || !strings.Contains(doctorOut, "big-thinker") {
		t.Fatalf("doctor must suggest migrating legacy keys to planner: %s", doctorOut)
	}

	startOut, code := runCent(t, bin, dir, "start", "another-feature")
	if code != 0 {
		t.Fatalf("start failed: %s", startOut)
	}
	if !strings.Contains(startOut, "planner") || !strings.Contains(startOut, "big-thinker") {
		t.Fatalf("start must print the same migration notice once: %s", startOut)
	}

	cmd := exec.Command(bin, "hook", "orchestration")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader("{}")
	hookOut, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hook orchestration failed: %v\n%s", err, hookOut)
	}
	if strings.Contains(string(hookOut), "retired plan role key") {
		t.Fatalf("hook orchestration must never repeat the migration notice: %s", hookOut)
	}
}
