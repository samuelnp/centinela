// Acceptance: specs/precommit-and-pr-gate.feature
package acceptance_test

import (
	"strings"
	"testing"
)

const prMarker = "<!-- centinela:pr-gate -->"

// Scenario: pr-gate emits a Markdown verdict listing each gate with a pass/fail marker and details
func TestPrGate_MarkdownVerdictWithDetails(t *testing.T) {
	skipWin(t)
	dir := pcRepo(t, "")
	pcBranch(t, dir, "internal/oversized.go", pcLines(140))
	out, code := runCent(t, buildCent(t), dir, "pr-gate")
	if code == 0 {
		t.Fatalf("a fail-severity gate must produce a non-zero exit\n%s", out)
	}
	if !strings.Contains(out, prMarker) || !strings.Contains(out, "| Gate | Status | Message |") {
		t.Fatalf("output must be the marked Markdown verdict table: %q", out)
	}
	if !strings.Contains(out, "❌") || !strings.Contains(out, "<details>") {
		t.Fatalf("failing gate must show a fail marker + details: %q", out)
	}
	if strings.Contains(strings.ToLower(out), "panic") {
		t.Fatalf("must not contain a stack trace: %q", out)
	}
}

// Scenario: pr-gate over an all-passing changeset exits 0 with a Markdown all-pass verdict
func TestPrGate_AllPassVerdict(t *testing.T) {
	skipWin(t)
	dir := pcRepo(t, "")
	pcBranch(t, dir, "internal/clean.go", pcLines(20))
	out, code := runCent(t, buildCent(t), dir, "pr-gate")
	if code != 0 {
		t.Fatalf("clean changeset must exit 0, exit=%d\n%s", code, out)
	}
	if !strings.Contains(out, prMarker) || !strings.Contains(out, "✅") {
		t.Fatalf("all-pass verdict must mark gates passing: %q", out)
	}
}

// Scenario: pr-gate run outside a PR context prints the verdict to stdout and does not post or error
func TestPrGate_OutsidePRContextPrintsToStdout(t *testing.T) {
	skipWin(t)
	dir := pcRepo(t, "")
	pcBranch(t, dir, "internal/clean.go", pcLines(20))
	out, code := runCent(t, buildCent(t), dir, "pr-gate")
	if code != 0 {
		t.Fatalf("clean changeset outside a PR must exit 0, exit=%d\n%s", code, out)
	}
	if !strings.Contains(out, prMarker) {
		t.Fatalf("verdict must still print to stdout outside a PR: %q", out)
	}
	if strings.Contains(strings.ToLower(out), "panic") || strings.Contains(strings.ToLower(out), "error:") {
		t.Fatalf("must not error or crash outside a PR: %q", out)
	}
}

// Scenario: fail_on_warning makes a failing warn gate fail the PR gate while the default does not
func TestPrGate_FailOnWarning(t *testing.T) {
	skipWin(t)
	bin := buildCent(t)
	warn := "\n[[gates.custom]]\nenabled = true\nname = \"style-nit\"\ncommand = \"false\"\nseverity = \"warn\"\n"

	def := pcRepo(t, warn) // [pr_gate] fail_on_warning omitted (default false)
	pcBranch(t, def, "internal/clean.go", pcLines(20))
	out, code := runCent(t, bin, def, "pr-gate")
	if code != 0 {
		t.Fatalf("default fail_on_warning=false: warn must not fail, exit=%d\n%s", code, out)
	}
	if !strings.Contains(out, "style-nit") || !strings.Contains(out, "⚠️") {
		t.Fatalf("warn gate must be reported as a warning: %q", out)
	}

	on := pcRepo(t, warn+"\n[pr_gate]\nfail_on_warning = true\n")
	pcBranch(t, on, "internal/clean.go", pcLines(20))
	out2, code2 := runCent(t, bin, on, "pr-gate")
	if code2 == 0 {
		t.Fatalf("fail_on_warning=true: a warn gate must fail the PR gate\n%s", out2)
	}
}

// Scenario: Custom gates and the audit-baseline gate participate in precommit and pr-gate like in validate
func TestPrGate_CustomAndAuditParticipate(t *testing.T) {
	skipWin(t)
	bin := buildCent(t)
	body := "\n[[gates.custom]]\nenabled = true\nname = \"no-todo\"\ncommand = \"false\"\nseverity = \"fail\"\n" +
		"\n[gates.audit_baseline]\nenabled = true\nseverity = \"fail\"\n"
	dir := pcRepo(t, body)
	writeFile(t, dir, "internal/clean.go", pcLines(20))
	pcGit(t, dir, "add", "internal/clean.go")
	pcOut, pcCode := runCent(t, bin, dir, "precommit")
	if pcCode == 0 || !strings.Contains(pcOut, "no-todo") {
		t.Fatalf("custom fail gate must block precommit and be named: exit=%d\n%s", pcCode, pcOut)
	}

	pcGit(t, dir, "commit", "-q", "-m", "stage")
	pcBranch(t, dir, "internal/clean2.go", pcLines(20))
	prOut, _ := runCent(t, bin, dir, "pr-gate")
	if !strings.Contains(prOut, "no-todo") {
		t.Fatalf("custom gate must appear in the pr-gate verdict: %q", prOut)
	}
}

// Scenario: Two runs over the same staged content produce identical verdict output and exit code
func TestPrGate_DeterministicAcrossRuns(t *testing.T) {
	skipWin(t)
	bin := buildCent(t)
	dir := pcRepo(t, "")
	pcBranch(t, dir, "internal/oversized.go", pcLines(140))
	out1, code1 := runCent(t, bin, dir, "pr-gate")
	out2, code2 := runCent(t, bin, dir, "pr-gate")
	if out1 != out2 {
		t.Fatalf("two runs over identical content must be byte-identical:\n---\n%s\n---\n%s", out1, out2)
	}
	if code1 != code2 || code1 == 0 {
		t.Fatalf("both runs must share the same non-zero exit, got %d and %d", code1, code2)
	}
}
