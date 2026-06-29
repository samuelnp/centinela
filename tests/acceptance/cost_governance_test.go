package acceptance_test

// Acceptance: specs/cost-governance.feature

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var costBinOnce sync.Once
var costBin string

func buildCostBin(t *testing.T) string {
	t.Helper()
	costBinOnce.Do(func() {
		dir, _ := os.MkdirTemp("", "cent-cost-bin")
		costBin = filepath.Join(dir, "centinela")
		c := exec.Command("go", "build", "-o", costBin, "./cmd/centinela")
		c.Dir = repoRoot(t)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("build: %v\n%s", err, out)
		}
	})
	return costBin
}

// seedCostAccRepo builds a temp repo: [cost] on, active workflow, a transcript.
func seedCostAccRepo(t *testing.T) (dir, transcript string) {
	t.Helper()
	dir = t.TempDir()
	w := func(rel, body string) {
		p := filepath.Join(dir, rel)
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	w("centinela.toml", "[cost]\nenabled=true\nstep_token_budget=1000\nfeature_token_budget=9000\n")
	w(".workflow/demo.json", `{"feature":"demo","currentStep":"code","stepOrder":["plan","code"],"steps":{}}`)
	transcript = filepath.Join(dir, "t.jsonl")
	w("t.jsonl", `{"message":{"usage":{"input_tokens":1500,"output_tokens":1000}}}`+"\n")
	return dir, transcript
}

// Scenario: capture from transcript, report over-budget, validate stays exit 0.
func TestAccCostCaptureReportAndSoftGate(t *testing.T) {
	bin := buildCostBin(t)
	dir, tp := seedCostAccRepo(t)

	hook := exec.Command(bin, "hook", "cost")
	hook.Dir = dir
	hook.Stdin = strings.NewReader(`{"cwd":"` + dir + `","transcript_path":"` + tp + `"}`)
	if out, err := hook.CombinedOutput(); err != nil {
		t.Fatalf("hook cost: %v\n%s", err, out)
	}

	report, code := runCent(t, bin, dir, "cost")
	if code != 0 || !strings.Contains(report, "OVER") || !strings.Contains(report, "demo/code") {
		t.Fatalf("cost report not over-budget (code=%d): %s", code, report)
	}

	out, code := runCent(t, bin, dir, "validate")
	if code != 0 {
		t.Fatalf("validate must stay exit 0 (soft gate), got %d: %s", code, out)
	}
	if !strings.Contains(out, "over budget") {
		t.Fatalf("validate should surface the cost ⚠: %s", out)
	}
}

// Scenario: missing transcript degrades gracefully (no error, no sample).
func TestAccCostMissingTranscriptNoOp(t *testing.T) {
	bin := buildCostBin(t)
	dir, _ := seedCostAccRepo(t)
	hook := exec.Command(bin, "hook", "cost")
	hook.Dir = dir
	hook.Stdin = strings.NewReader(`{"cwd":"` + dir + `"}`)
	if out, err := hook.CombinedOutput(); err != nil {
		t.Fatalf("missing transcript must not error: %v\n%s", err, out)
	}
	if _, err := os.Stat(filepath.Join(dir, ".workflow/telemetry/events.jsonl")); err == nil {
		t.Fatal("no transcript_path should record nothing")
	}
}
