// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import (
	"path/filepath"
	"strings"
	"testing"
)

// rshFreshFixture is a real git repo with a source file and roadmap state,
// ready for `artifact stamp` + `complete` (mirrors avvFixture).
func rshFreshFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "config", "user.email", "rsh@centinela.dev")
	runGit(t, dir, "config", "user.name", "RSH")
	mustWrite(t, filepath.Join(dir, "src.go"), "package x\n")
	mustWrite(t, filepath.Join(dir, ".workflow", "roadmap.json"), rshBaseRoadmap)
	commit(t, dir, "baseline")
	return dir
}

// Scenario: a deferral after the verification stamp does not stale the verification
func TestRsh_DeferralAfterStampStaysFresh(t *testing.T) {
	bin := buildCent(t)
	dir := rshFreshFixture(t)
	avvSeedWorkflow(t, dir, "late-finding", "adversarial-v1")
	avvWriteReport(t, dir, "late-finding", avvReport("SAFE", "", avvGroundedCommands))
	avvStamp(t, bin, dir, "late-finding")

	out, code := runCent(t, bin, dir, "roadmap", "defer", "late-thing", "--summary", "x")
	if code != 0 {
		t.Fatalf("defer exit=%d\n%s", code, out)
	}

	cout, ccode := avvComplete(t, bin, dir, "late-finding")
	if ccode != 0 || strings.Contains(cout, "stale") {
		t.Fatalf("a roadmap-state-only commit after the stamp must stay fresh: %d\n%s", ccode, cout)
	}
}

// Scenario: an uncommitted regenerated ROADMAP.md does not stale the verification
func TestRsh_UncommittedRegeneratedMarkdownStaysFresh(t *testing.T) {
	bin := buildCent(t)
	dir := rshFreshFixture(t)
	mustWrite(t, filepath.Join(dir, "centinela.toml"), "[workflow]\ndisable_auto_commit = true\n")
	avvSeedWorkflow(t, dir, "policy-thing", "adversarial-v1")
	avvWriteReport(t, dir, "policy-thing", avvReport("SAFE", "", avvGroundedCommands))
	avvStamp(t, bin, dir, "policy-thing")

	out, code := runCent(t, bin, dir, "roadmap", "defer", "another-thing", "--summary", "x")
	if code != 0 {
		t.Fatalf("defer exit=%d\n%s", code, out)
	}

	cout, ccode := avvComplete(t, bin, dir, "policy-thing")
	if ccode != 0 || strings.Contains(cout, "stale") {
		t.Fatalf("uncommitted regenerated roadmap state must stay fresh: %d\n%s", ccode, cout)
	}
}

// Scenario: a source change committed after the stamp still stales the verification
func TestRsh_SourceChangeAfterStampStillStales(t *testing.T) {
	bin := buildCent(t)
	dir := rshFreshFixture(t)
	avvSeedWorkflow(t, dir, "mixed-range", "adversarial-v1")
	avvWriteReport(t, dir, "mixed-range", avvReport("SAFE", "", avvGroundedCommands))
	avvStamp(t, bin, dir, "mixed-range")

	if out, code := runCent(t, bin, dir, "roadmap", "defer", "x", "--summary", "x"); code != 0 {
		t.Fatalf("defer exit=%d\n%s", code, out)
	}
	mustWrite(t, filepath.Join(dir, "src.go"), "package x // edited\n")
	commit(t, dir, "feat: real change")

	out, code := avvComplete(t, bin, dir, "mixed-range")
	if code == 0 {
		t.Fatalf("a source commit in the same range must stale the verification, got exit 0: %s", out)
	}
	containsAll(t, out, "stale")
}

// Scenario: an unreadable revision range fails closed
func TestRsh_UnreadableRevisionRangeFailsClosed(t *testing.T) {
	bin := buildCent(t)
	dir := rshFreshFixture(t)
	avvSeedWorkflow(t, dir, "bogus-rev", "adversarial-v1")
	bogus := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	report := "### Adversarial Verifier Report: demo\n**Status:** SAFE\n\n#### Findings\n\n\n" +
		"```json centinela:verification\n{\"revision\":\"" + bogus + "\",\"treeDigest\":\"irrelevant\"," +
		"\"commands\":" + avvGroundedCommands + "}\n```\n"
	avvWriteReport(t, dir, "bogus-rev", report)

	out, code := avvComplete(t, bin, dir, "bogus-rev")
	if code == 0 {
		t.Fatalf("an unresolvable stamped revision must fail closed as stale, got exit 0: %s", out)
	}
	containsAll(t, out, "stale")
}
