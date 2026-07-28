// Acceptance: specs/adversarial-validate-verifier.feature
package acceptance_test

import (
	"strings"
	"testing"
)

// Scenario: A report without a non-empty commands-run record fails evidence validation
func TestAVV_EmptyCommandsRecordBlocks(t *testing.T) {
	bin := avvBuildBin(t)
	dir := avvFixture(t)
	avvSeedWorkflow(t, dir, "dead-subagent-stub", "adversarial-v1")
	avvWriteReport(t, dir, "dead-subagent-stub", avvReport("SAFE", "", "[]"))

	out, code := avvComplete(t, bin, dir, "dead-subagent-stub")
	if code == 0 {
		t.Fatalf("empty commands array must block, got exit 0: %s", out)
	}
	if !strings.Contains(out, "no commands-run record") {
		t.Fatalf("message must say no commands-run record: %s", out)
	}
	if !strings.Contains(out, "gatekeeper-prompt.md") {
		t.Fatalf("message must name the prompt doc as the remedy: %s", out)
	}
}

// Scenario: centinela artifact new followed immediately by centinela complete FAILS
func TestAVV_ArtifactNewStubThenCompleteFails(t *testing.T) {
	bin := avvBuildBin(t)
	dir := avvFixture(t)
	avvSeedWorkflow(t, dir, "just-scaffolded", "adversarial-v1")
	if out, code := runCent(t, bin, dir, "artifact", "new", "just-scaffolded", "gatekeeper"); code != 0 {
		t.Fatalf("artifact new must succeed: %s", out)
	}

	out, code := avvComplete(t, bin, dir, "just-scaffolded")
	if code == 0 {
		t.Fatalf("a freshly scaffolded stub must never satisfy complete, got exit 0: %s", out)
	}
	if !strings.Contains(out, "no commands-run record") {
		t.Fatalf("message must say no commands-run record: %s", out)
	}
}

// Scenario: A report whose commands never include a passing "centinela validate" run is refused
func TestAVV_PartialCommandsWithoutPassingValidateRefused(t *testing.T) {
	bin := avvBuildBin(t)
	dir := avvFixture(t)
	avvSeedWorkflow(t, dir, "partial-commands", "adversarial-v1")
	commands := `[{"argv":["go","test","./..."],"exitCode":0,"durationMs":5000}]`
	avvWriteReport(t, dir, "partial-commands", avvReport("SAFE", "", commands))

	out, code := avvComplete(t, bin, dir, "partial-commands")
	if code == 0 {
		t.Fatalf("missing a passing 'centinela validate' entry must block, got exit 0: %s", out)
	}
	if !strings.Contains(out, "no commands-run record") {
		t.Fatalf("message must say no commands-run record: %s", out)
	}
}

// Scenario: A verifier that cannot execute commands in its harness fails closed
func TestAVV_NoBashHarnessFailsClosed(t *testing.T) {
	bin := avvBuildBin(t)
	dir := avvFixture(t)
	avvSeedWorkflow(t, dir, "no-bash-harness", "adversarial-v1")
	report := "### Adversarial Verifier Report: no-bash-harness\n**Status:** CRITICAL\n\n" +
		"#### Commands Run\n- could not execute commands in this harness\n\n" +
		"```json centinela:verification\n{\"revision\":\"\",\"treeDigest\":\"\",\"commands\":[]}\n```\n"
	avvWriteReport(t, dir, "no-bash-harness", report)

	out, code := avvComplete(t, bin, dir, "no-bash-harness")
	if code == 0 {
		t.Fatalf("a harness that cannot execute commands must never narrate a pass, got exit 0: %s", out)
	}
	if !strings.Contains(out, "CRITICAL") {
		t.Fatalf("the fail-closed CRITICAL verdict must be the block reason: %s", out)
	}
}
