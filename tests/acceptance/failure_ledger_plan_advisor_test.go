// Acceptance: specs/failure-ledger-plan-advisor.feature
package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/insights"
	"github.com/samuelnp/centinela/internal/telemetry"
	"github.com/samuelnp/centinela/internal/workflow"
)

// flFullBrief covers every generic advisor-question topic so the default
// question cap is not exhausted before the gate-failure pre-warning question.
const flFullBrief = "## Problem\ntext\n## Scope\nin scope\n## Constraints\nsecurity\n" +
	"## Risks\ntradeoff\n## Acceptance Criteria\nGiven when then\n## Edge Cases\ninvalid input\n"

// flBuildBinary builds the centinela binary once per test, returning its path.
func flBuildBinary(t *testing.T) string {
	t.Helper()
	o, _ := os.Getwd()
	repo := filepath.Clean(filepath.Join(o, "..", ".."))
	bin := filepath.Join(t.TempDir(), "centinela-fl")
	build := exec.Command("go", "build", "-o", bin, "./cmd/centinela")
	build.Dir = repo
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build centinela failed: %v\n%s", err, out)
	}
	return bin
}

// flRepo lays out a temp repo with a plan-step workflow named "f", a feature
// brief, a roadmap, and (optionally) a telemetry ledger. It returns the repo dir.
// ledger=="" means no ledger file; toml is written verbatim when non-empty.
func flRepo(t *testing.T, brief, ledger, toml string) string {
	t.Helper()
	d := t.TempDir()
	o, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(o) }) //nolint:errcheck
	if err := os.Chdir(d); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	os.MkdirAll(workflow.WorkflowDir, 0o755)                                                                      //nolint:errcheck
	os.MkdirAll("docs/features", 0o755)                                                                           //nolint:errcheck
	os.MkdirAll(filepath.Join(".workflow", "telemetry"), 0o755)                                                   //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte(brief), 0o644)                                                      //nolint:errcheck
	os.WriteFile(".workflow/roadmap.json", []byte(`{"phases":[{"name":"P1","features":[{"name":"f"}]}]}`), 0o644) //nolint:errcheck
	if toml != "" {
		os.WriteFile("centinela.toml", []byte(toml), 0o644) //nolint:errcheck
	}
	if ledger != "" {
		os.WriteFile(filepath.Join(".workflow", "telemetry", "events.jsonl"), []byte(ledger), 0o644) //nolint:errcheck
	}
	if err := workflow.Save(workflow.New("f")); err != nil {
		t.Fatalf("save workflow: %v", err)
	}
	return d
}

// flRun runs `centinela hook plan-advisor` in dir and returns combined output.
func flRun(t *testing.T, bin, dir string) string {
	t.Helper()
	cmd := exec.Command(bin, "hook", "plan-advisor")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("plan-advisor hook failed: %v\n%s", err, out)
	}
	return string(out)
}

const flRankedLedger = `{"type":"gate-failure","gate":"g1-file-size"}
{"type":"gate-failure","gate":"g1-file-size"}
{"type":"gate-failure","gate":"g1-file-size"}
{"type":"gate-failure","gate":"coverage"}
{"type":"gate-failure","gate":"coverage"}
{"type":"gate-failure","gate":"import-graph"}
`

// Scenario: Missing ledger file leaves advisor output byte-identical to today
func TestFL_MissingLedgerByteIdentical(t *testing.T) {
	bin := flBuildBinary(t)
	missing := flRun(t, bin, flRepo(t, "## Problem\ntext\n", "", ""))
	disabled := flRun(t, bin, flRepo(t, "## Problem\ntext\n", flRankedLedger, "[telemetry]\nenabled = false\n"))
	if strings.Contains(missing, "recurring gate failures") {
		t.Fatalf("missing ledger emitted a recurring-failure line:\n%s", missing)
	}
	if strings.Contains(missing, "worst:") {
		t.Fatalf("missing ledger emitted a pre-warning question:\n%s", missing)
	}
	if missing != disabled {
		t.Fatalf("missing-ledger output must equal feature-disabled output\nmissing:\n%s\ndisabled:\n%s", missing, disabled)
	}
}

// Scenario: Empty ledger file leaves advisor output byte-identical to today
func TestFL_EmptyLedgerNoFailureOutput(t *testing.T) {
	bin := flBuildBinary(t)
	dir := flRepo(t, "## Problem\ntext\n", "", "")
	os.WriteFile(filepath.Join(dir, ".workflow", "telemetry", "events.jsonl"), []byte(""), 0o644) //nolint:errcheck
	out := flRun(t, bin, dir)
	if strings.Contains(out, "recurring gate failures") || strings.Contains(out, "worst:") {
		t.Fatalf("empty ledger must produce no failure output:\n%s", out)
	}
}

// Scenario: Ledger with only block and step-advanced events produces no recurring-failure output
func TestFL_NonFailureEventsProduceNothing(t *testing.T) {
	bin := flBuildBinary(t)
	ledger := `{"type":"block","reason":"out-of-step","fileType":"plan","feature":"alpha"}
{"type":"block","reason":"need-init","fileType":"source","feature":"alpha"}
{"type":"step-advanced","feature":"alpha"}
`
	out := flRun(t, bin, flRepo(t, "## Problem\ntext\n", ledger, ""))
	if strings.Contains(out, "recurring gate failures") || strings.Contains(out, "worst:") {
		t.Fatalf("non-gate-failure events must produce no failure output:\n%s", out)
	}
}

// Scenario: Telemetry disabled in config suppresses all ledger-derived failure context
func TestFL_TelemetryDisabledSuppressesContext(t *testing.T) {
	bin := flBuildBinary(t)
	disabled := flRun(t, bin, flRepo(t, "## Problem\ntext\n", flRankedLedger, "[telemetry]\nenabled = false\n"))
	missing := flRun(t, bin, flRepo(t, "## Problem\ntext\n", "", ""))
	if strings.Contains(disabled, "recurring gate failures") || strings.Contains(disabled, "worst:") {
		t.Fatalf("disabled telemetry must suppress failure context:\n%s", disabled)
	}
	if disabled != missing {
		t.Fatalf("disabled output must equal missing-ledger output\ndisabled:\n%s\nmissing:\n%s", disabled, missing)
	}
}

// Scenario: Recurring gate failures appear in the context summary ranked by count descending
func TestFL_RecurringFailuresRankedDescending(t *testing.T) {
	bin := flBuildBinary(t)
	out := flRun(t, bin, flRepo(t, "## Problem\ntext\n", flRankedLedger, ""))
	want := "- recurring gate failures: g1-file-size (×3), coverage (×2), import-graph (×1)"
	if !strings.Contains(out, want) {
		t.Fatalf("expected ranked recurring-failure line %q, got:\n%s", want, out)
	}
	if !strings.Contains(out, "Relevant context:") {
		t.Fatalf("failure line must live under the Relevant context block:\n%s", out)
	}
}

// Scenario: Recurring gate failures counts match centinela insights for the same ledger
func TestFL_CountsMatchInsights(t *testing.T) {
	bin := flBuildBinary(t)
	out := flRun(t, bin, flRepo(t, "## Problem\ntext\n", flRankedLedger, ""))
	events, err := telemetry.ReadDefault()
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	for _, g := range insights.Gates(events, 3) {
		entry := g.Key + " (×" + itoa(g.Count) + ")"
		if !strings.Contains(out, entry) {
			t.Fatalf("advisor output missing insights entry %q:\n%s", entry, out)
		}
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// Scenario: Ties in failure count break by gate name ascending for reproducible output
func TestFL_TiesBreakByNameAscending(t *testing.T) {
	bin := flBuildBinary(t)
	ledger := `{"type":"gate-failure","gate":"z-gate"}
{"type":"gate-failure","gate":"z-gate"}
{"type":"gate-failure","gate":"a-gate"}
{"type":"gate-failure","gate":"a-gate"}
{"type":"gate-failure","gate":"m-gate"}
{"type":"gate-failure","gate":"m-gate"}
`
	out := flRun(t, bin, flRepo(t, "## Problem\ntext\n", ledger, ""))
	want := "- recurring gate failures: a-gate (×2), m-gate (×2), z-gate (×2)"
	if !strings.Contains(out, want) {
		t.Fatalf("expected tie-break order %q, got:\n%s", want, out)
	}
}

// Scenario: A gate-failure event with an empty Gate field buckets under "<none>" without crashing
func TestFL_EmptyGateBucketsAsNone(t *testing.T) {
	bin := flBuildBinary(t)
	ledger := `{"type":"gate-failure","gate":""}
{"type":"gate-failure"}
`
	out := flRun(t, bin, flRepo(t, "## Problem\ntext\n", ledger, ""))
	if !strings.Contains(out, "<none> (×2)") {
		t.Fatalf("empty Gate must bucket under <none>:\n%s", out)
	}
}

// Scenario: Only the top-N gates are listed when more distinct gates failed
func TestFL_OnlyTopNListed(t *testing.T) {
	bin := flBuildBinary(t)
	var sb strings.Builder
	// 8 distinct buckets, gate-i failing i+1 times (g-7 highest .. g-0 lowest).
	gates := []string{"g-0", "g-1", "g-2", "g-3", "g-4", "g-5", "g-6", "g-7"}
	for i, g := range gates {
		for j := 0; j <= i; j++ {
			sb.WriteString(`{"type":"gate-failure","gate":"` + g + `"}` + "\n")
		}
	}
	out := flRun(t, bin, flRepo(t, "## Problem\ntext\n", sb.String(), "[workflow]\nplan_advisor_failure_top_n = 3\n"))
	want := "- recurring gate failures: g-7 (×8), g-6 (×7), g-5 (×6)"
	if !strings.Contains(out, want) {
		t.Fatalf("expected exactly the top-3 gates %q, got:\n%s", want, out)
	}
	if strings.Contains(out, "g-4 (×5)") {
		t.Fatalf("top-N must truncate beyond 3 gates:\n%s", out)
	}
}

// Scenario: A gate recurring at or above threshold produces a pre-warning question naming the gate
func TestFL_PreWarningQuestionNamesGate(t *testing.T) {
	bin := flBuildBinary(t)
	ledger := strings.Repeat(`{"type":"gate-failure","gate":"g1-file-size"}`+"\n", 5)
	out := flRun(t, bin, flRepo(t, flFullBrief, ledger, ""))
	if !strings.Contains(out, "recurring gate failures (worst: g1-file-size)") {
		t.Fatalf("expected pre-warning question naming g1-file-size:\n%s", out)
	}
	if !strings.Contains(out, "[feature-specialist]") {
		t.Fatalf("pre-warning question must carry a lens tag:\n%s", out)
	}
}

// Scenario: A gate below the recurrence threshold produces no pre-warning question
func TestFL_BelowThresholdNoQuestion(t *testing.T) {
	bin := flBuildBinary(t)
	// Every gate fails at most 2 times — below the recurrence threshold of 3.
	ledger := strings.Repeat(`{"type":"gate-failure","gate":"g1-file-size"}`+"\n", 2) +
		strings.Repeat(`{"type":"gate-failure","gate":"coverage"}`+"\n", 2)
	out := flRun(t, bin, flRepo(t, flFullBrief, ledger, ""))
	if strings.Contains(out, "worst:") {
		t.Fatalf("gates at count 2 (below threshold 3) must not produce a pre-warning question:\n%s", out)
	}
}

// Scenario: The pre-warning question respects the plan_question_limit cap
func TestFL_RespectsQuestionLimit(t *testing.T) {
	bin := flBuildBinary(t)
	// Bare brief leaves all generic questions live; cap = 3 must hold total ≤ 3.
	ledger := strings.Repeat(`{"type":"gate-failure","gate":"coverage"}`+"\n", 4)
	out := flRun(t, bin, flRepo(t, "(no headings)\n", ledger, "[workflow]\nplan_question_limit = 3\n"))
	if n := strings.Count(out, "- ["); n > 3 {
		t.Fatalf("question count %d exceeds plan_question_limit=3:\n%s", n, out)
	}
}

// Scenario: The advisor only surfaces recurring failures during the plan step
func TestFL_OnlyDuringPlanStep(t *testing.T) {
	bin := flBuildBinary(t)
	dir := flRepo(t, "## Problem\ntext\n", flRankedLedger, "")
	wf := workflow.New("f")
	wf.CurrentStep = "code"
	if err := workflow.Save(wf); err != nil {
		t.Fatalf("save code-step workflow: %v", err)
	}
	out := flRun(t, bin, dir)
	if strings.Contains(out, "recurring gate failures") || strings.Contains(out, "worst:") {
		t.Fatalf("advisor must stay silent outside the plan step:\n%s", out)
	}
}

// Scenario: Headless mode leaves advisor behaviour unchanged
func TestFL_HeadlessUnchanged(t *testing.T) {
	bin := flBuildBinary(t)
	dir := flRepo(t, "## Problem\ntext\n", flRankedLedger, "")
	cmd := exec.Command(bin, "hook", "plan-advisor")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "CENTINELA_HEADLESS=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("headless hook failed: %v\n%s", err, out)
	}
	if strings.TrimSpace(string(out)) != "" {
		t.Fatalf("headless advisor must emit nothing, got:\n%s", out)
	}
}

// Scenario: The advisor never writes to the ledger
func TestFL_NeverWritesLedger(t *testing.T) {
	bin := flBuildBinary(t)
	dir := flRepo(t, "## Problem\ntext\n", flRankedLedger, "")
	path := filepath.Join(dir, ".workflow", "telemetry", "events.jsonl")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read before: %v", err)
	}
	flRun(t, bin, dir)
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read after: %v", err)
	}
	if string(before) != string(after) {
		t.Fatalf("advisor must not modify the ledger\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// Scenario: Two runs on the same ledger produce byte-identical advisor output
func TestFL_DeterministicOutput(t *testing.T) {
	bin := flBuildBinary(t)
	dir := flRepo(t, "## Problem\ntext\n", flRankedLedger, "")
	first := flRun(t, bin, dir)
	second := flRun(t, bin, dir)
	if first != second {
		t.Fatalf("two runs must be byte-identical\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}
