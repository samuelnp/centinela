// Acceptance: specs/adversarial-validate-verifier.feature
package acceptance_test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Scenario: The validate-step hook directive names the adversarial verifier
// and its reasoning tier (AC6). Strict orchestration mode is required for
// `hook orchestration` to emit a directive at all.
func TestAVV_HookDirectiveNamesVerifierAndReasoningTier(t *testing.T) {
	bin := avvBuildBin(t)
	dir := avvFixture(t)
	body := `{"feature":"directive-check","currentStep":"validate",` +
		`"stepOrder":["plan","code","tests","validate","docs"],"steps":{},` +
		`"validateContract":"adversarial-v1","orchestrationMode":"strict-subagents-v1"}`
	mustWrite(t, filepath.Join(dir, ".workflow", "directive-check.json"), body)

	cmd := exec.Command(bin, "hook", "orchestration")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader("{}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hook orchestration failed: %v\n%s", err, out)
	}
	got := string(out)
	for _, want := range []string{"gatekeeper", "reasoning", "file paths", "do NOT summarize"} {
		if !strings.Contains(got, want) {
			t.Fatalf("directive missing %q: %s", want, got)
		}
	}
}
