package acceptance_test

// Acceptance: specs/configurable-model-routing.feature (AC1, AC2, AC7, edges).

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// codeStepRepo sets up a repo whose active workflow is on the code step (so
// senior-engineer is a required role) with the given centinela.toml body.
func codeStepRepo(t *testing.T, toml string) (dir, bin string) {
	t.Helper()
	bin = buildModelsTestBinary(t, t.TempDir())
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) })    //nolint:errcheck
	os.Chdir(d)                             //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck
	wf := workflow.New("myfeature")
	wf.CurrentStep = "code"
	workflow.Save(wf)                                  //nolint:errcheck
	os.WriteFile("centinela.toml", []byte(toml), 0644) //nolint:errcheck
	return d, bin
}

// AC1: tier remap — opencode reasoning resolves to the configured Kimi model.
func TestRouting_TierRemapForRunner(t *testing.T) {
	d, bin := setupModelsRepo(t, "[orchestration.model_map.reasoning]\nopencode = \"moonshotai/kimi-k2\"\n")
	out, _ := runBin(t, bin, d, "hook", "orchestration")
	if !strings.Contains(out, "model: moonshotai/kimi-k2 (opencode)") {
		t.Errorf("AC1: expected kimi for opencode reasoning; got:\n%s", out)
	}
}

// AC2: role override beats the role's tier for the active runner.
func TestRouting_RoleOverrideBeatsTier(t *testing.T) {
	d, bin := codeStepRepo(t, "[orchestration.models]\nsenior-engineer = { opencode = \"deepseek/deepseek-coder\" }\n")
	out, _ := runBin(t, bin, d, "hook", "orchestration")
	if !strings.Contains(out, "model: deepseek/deepseek-coder (opencode)") {
		t.Errorf("AC2: expected override for senior-engineer; got:\n%s", out)
	}
	if !strings.Contains(out, "senior-engineer (model: claude-opus-4-7 (claude)") {
		t.Errorf("AC2: claude column should still show its default; got:\n%s", out)
	}
}

// AC7: codex (empty column) carries the tier name, never another runner's ID.
func TestRouting_CodexRule4NoLeak(t *testing.T) {
	d, bin := setupModelsRepo(t, "[orchestration.model_map.reasoning]\nopencode = \"moonshotai/kimi-k2\"\n")
	out, _ := runBin(t, bin, d, "hook", "orchestration")
	if !strings.Contains(out, "model: reasoning (codex)") {
		t.Errorf("AC7: expected codex tier-name fallback; got:\n%s", out)
	}
	if strings.Contains(out, "moonshotai/kimi-k2 (codex)") {
		t.Errorf("AC7: codex must not carry the opencode ID; got:\n%s", out)
	}
}

// Edge: role override wins over a model_map entry for the same runner+tier.
func TestRouting_OverrideBeatsModelMap(t *testing.T) {
	toml := "[orchestration.model_map.reasoning]\nopencode = \"moonshotai/kimi-k2\"\n" +
		"[orchestration.models]\nbig-thinker = { opencode = \"deepseek/deepseek-coder\" }\n"
	d, bin := setupModelsRepo(t, toml)
	out, _ := runBin(t, bin, d, "hook", "orchestration")
	if !strings.Contains(out, "big-thinker (model: claude-opus-4-7 (claude), model: deepseek/deepseek-coder (opencode)") {
		t.Errorf("edge: override should beat model_map; got:\n%s", out)
	}
}

// Edge: empty tables behave like absent tables — built-in defaults everywhere.
func TestRouting_EmptyTablesDefault(t *testing.T) {
	d, bin := setupModelsRepo(t, "[orchestration.model_map]\n[orchestration.models]\n")
	out, _ := runBin(t, bin, d, "hook", "orchestration")
	if !strings.Contains(out, "big-thinker (model: claude-opus-4-7 (claude)") {
		t.Errorf("edge: empty tables should default; got:\n%s", out)
	}
}
