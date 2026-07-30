// Acceptance: specs/docs-step-markdown-first.feature
package acceptance_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// dsmfRepo provisions a repo with the docs-context inputs for feature "f".
func dsmfRepo(t *testing.T, withPlan, withSpec, withChangelog bool) string {
	t.Helper()
	dir := t.TempDir()
	mustWrite(t, dir+"/docs/features/f.md", "# f brief body\n")
	if withPlan {
		mustWrite(t, dir+"/docs/plans/f.md", "# f plan body\n")
	}
	if withSpec {
		mustWrite(t, dir+"/specs/f.feature", "Feature: f\n  Scenario: s\n")
	}
	if withChangelog {
		mustWrite(t, dir+"/"+workflow.WorkflowDir+"/f-changelog.md", "- feat: f draft\n")
	}
	return dir
}

// Scenario: docs context prints the curated feature-scale inputs
func TestDSMFDocsContextPrintsCuratedInputs(t *testing.T) {
	bin := buildCent(t)
	dir := dsmfRepo(t, true, true, true)
	out, code := runCent(t, bin, dir, "docs", "context", "f")
	if code != 0 {
		t.Fatalf("docs context must exit 0 (code=%d):\n%s", code, out)
	}
	for _, want := range []string{
		"## Feature brief", "f brief body",
		"## Plan", "f plan body",
		"## Spec scenarios", "Feature: f",
		"## Changelog draft", "- feat: f draft",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdout missing %q:\n%s", want, out)
		}
	}
}

// Scenario: docs context reports every missing required input at once
func TestDSMFDocsContextAggregatesMissingInputs(t *testing.T) {
	bin := buildCent(t)
	dir := dsmfRepo(t, false, false, false)
	out, code := runCent(t, bin, dir, "docs", "context", "f")
	if code == 0 {
		t.Fatalf("docs context must exit non-zero with missing inputs:\n%s", out)
	}
	for _, want := range []string{"docs/plans/f.md", "specs/f.feature"} {
		if !strings.Contains(out, want) {
			t.Fatalf("error must name %q:\n%s", want, out)
		}
	}
}

// Scenario: docs context treats the changelog draft as optional
func TestDSMFDocsContextChangelogOptionalHint(t *testing.T) {
	bin := buildCent(t)
	dir := dsmfRepo(t, true, true, false)
	out, code := runCent(t, bin, dir, "docs", "context", "f")
	if code != 0 {
		t.Fatalf("absent changelog draft must keep exit 0 (code=%d):\n%s", code, out)
	}
	if !strings.Contains(out, "centinela artifact new f changelog") {
		t.Fatalf("changelog section must suggest the artifact-new command:\n%s", out)
	}
}

// Scenario: docs context prints the curated feature-scale inputs
// (idempotency: the command is read-only — two runs are byte-identical)
func TestDSMFDocsContextIdempotentDoubleRun(t *testing.T) {
	bin := buildCent(t)
	dir := dsmfRepo(t, true, true, true)
	first, code1 := runCent(t, bin, dir, "docs", "context", "f")
	second, code2 := runCent(t, bin, dir, "docs", "context", "f")
	if code1 != 0 || code2 != 0 {
		t.Fatalf("both runs must exit 0 (codes %d/%d)", code1, code2)
	}
	if first != second {
		t.Fatalf("docs context must be deterministic:\n--- first\n%s\n--- second\n%s", first, second)
	}
}
