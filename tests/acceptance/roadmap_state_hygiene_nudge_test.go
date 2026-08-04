// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import (
	"path/filepath"
	"testing"
)

// rshNudgeRepo builds a repo with a roadmap holding one schedulable feature
// per name plus backlogN Backlog findings (aged 20 days, so 20 > 14 stale).
func rshNudgeRepo(t *testing.T, names []string, backlogN int) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "main")
	feats := ""
	for i, n := range names {
		if i > 0 {
			feats += ","
		}
		feats += `{"name":"` + n + `"}`
	}
	backlog := ""
	for i := 0; i < backlogN; i++ {
		if i > 0 {
			backlog += ","
		}
		backlog += `{"name":"deferred-` + string(rune('a'+i)) +
			`","summary":"s","deferredAt":"2026-01-01T00:00:00Z"}`
	}
	rm := `{"phases":[{"name":"P","features":[` + feats + `]},` +
		`{"name":"Backlog","features":[` + backlog + `]}]}`
	mustWrite(t, filepath.Join(dir, ".workflow", "roadmap.json"), rm)
	mustWrite(t, filepath.Join(dir, "centinela.toml"),
		"[workflow]\ndisable_auto_commit=true\n")
	return dir
}

// rshDocsWorkflow seeds feature at its final "docs" step, ready to complete to done.
func rshDocsWorkflow(t *testing.T, dir, feature string) {
	t.Helper()
	steps := `{"plan":{"status":"done"},"code":{"status":"done"},"tests":{"status":"done"},` +
		`"validate":{"status":"done"},"docs":{"status":"in-progress"}}`
	mustWrite(t, filepath.Join(dir, ".workflow", feature+".json"),
		`{"feature":"`+feature+`","currentStep":"docs","steps":`+steps+`}`)
	mustWrite(t, filepath.Join(dir, ".workflow", feature+"-changelog.md"), "- feat: "+feature+"\n")
}

// Scenario: completing the last schedulable feature nudges about the Backlog
func TestRsh_CompletingLastFeatureNudgesAboutBacklog(t *testing.T) {
	bin := buildCent(t)
	dir := rshNudgeRepo(t, []string{"alpha", "last-one"}, 12)
	seedDoneAt(t, dir, "alpha")
	rshDocsWorkflow(t, dir, "last-one")

	out, code := runCent(t, bin, dir, "complete", "last-one")
	if code != 0 {
		t.Fatalf("exit=%d\n%s", code, out)
	}
	containsAll(t, out, "Roadmap complete", "12 deferred findings",
		"roadmap backlog --stale", "roadmap promote")
}

// Scenario: no nudge while schedulable work remains
func TestRsh_NoNudgeWhileScheduleableWorkRemains(t *testing.T) {
	bin := buildCent(t)
	dir := rshNudgeRepo(t, []string{"alpha", "last-one"}, 3)
	rshDocsWorkflow(t, dir, "last-one") // "alpha" stays incomplete

	out, code := runCent(t, bin, dir, "complete", "last-one")
	if code != 0 {
		t.Fatalf("exit=%d\n%s", code, out)
	}
	if contains(out, "deferred findings remain") {
		t.Fatalf("nudge must not print while schedulable work remains:\n%s", out)
	}
}
