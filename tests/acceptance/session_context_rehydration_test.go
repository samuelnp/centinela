package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const sessionDirective = "CENTINELA DIRECTIVE: session rehydration"

// newBin builds the CLI into a fresh temp dir. (buildCentinela/mustContain/
// mustNotContain are shared acceptance helpers defined elsewhere in the package.)
func newBin(t *testing.T) string {
	t.Helper()
	return buildCentinela(t, t.TempDir())
}

// run executes a centinela hook subcommand inside dir with the given stdin.
func run(t *testing.T, bin, dir, stdin string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(stdin)
	out, err := cmd.CombinedOutput()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("run %v: %v", args, err)
	}
	return string(out), code
}

func wfDir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, ".workflow"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func write(t *testing.T, dir, rel, body string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func doneWF(t *testing.T, dir, feature string) {
	write(t, dir, ".workflow/"+feature+".json",
		`{"feature":"`+feature+`","currentStep":"done","steps":{}}`)
}

// Scenario Outline (startup|clear|compact|resume): identical rehydration payload
// regardless of SessionStart source.
func TestAcceptance_SessionStartPayloadOnEachSource(t *testing.T) {
	bin := newBin(t)
	for _, src := range []string{"startup", "clear", "compact", "resume"} {
		dir := t.TempDir()
		write(t, dir, "PROJECT.md", "x")
		write(t, dir, ".workflow/roadmap.json",
			`{"phases":[{"name":"Phase 0","features":[{"name":"next-feature"},{"name":"later"}]}]}`)
		out, code := run(t, bin, dir, `{"source":"`+src+`"}`, "hook", "session")
		if code != 0 {
			t.Fatalf("source %s: exit %d\n%s", src, code, out)
		}
		mustContain(t, out, sessionDirective)
		mustContain(t, out, "next-feature") // full roadmap + next feature
		mustContain(t, out, "(planned)")    // per-feature status
		mustContain(t, out, "PROJECT.md")   // pointer path
		mustContain(t, out, "docs/features/next-feature.md")
		mustNotContain(t, out, "## Problem") // paths only — no inlined brief contents
	}
}

// Next feature is the first incomplete across ALL phases, not just Phase 0.
func TestAcceptance_NextIsFirstIncompleteAcrossPhases(t *testing.T) {
	bin := newBin(t)
	dir := t.TempDir()
	write(t, dir, "PROJECT.md", "x")
	write(t, dir, ".workflow/roadmap.json",
		`{"phases":[{"name":"Phase 0","features":[{"name":"p0a"},{"name":"p0b"}]},`+
			`{"name":"Phase 1","features":[{"name":"phase-1-first"}]}]}`)
	doneWF(t, dir, "p0a")
	doneWF(t, dir, "p0b")
	out, code := run(t, bin, dir, `{"source":"clear"}`, "hook", "session")
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	mustContain(t, out, "Next feature to plan: phase-1-first")
	mustContain(t, out, "docs/features/phase-1-first.md")
}

// Every feature done → roadmap-complete, no next name, no <next>.md pointer.
func TestAcceptance_AllDoneRoadmapComplete(t *testing.T) {
	bin := newBin(t)
	dir := t.TempDir()
	write(t, dir, "PROJECT.md", "x")
	write(t, dir, ".workflow/roadmap.json",
		`{"phases":[{"name":"P","features":[{"name":"only"}]}]}`)
	doneWF(t, dir, "only")
	out, code := run(t, bin, dir, "{}", "hook", "session")
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	mustContain(t, out, "Roadmap complete")
	mustNotContain(t, out, "Next feature to plan:")
	mustNotContain(t, out, "docs/features/")
}

// Missing roadmap and malformed roadmap json both exit zero with no payload.
func TestAcceptance_MissingAndInvalidRoadmapSilent(t *testing.T) {
	bin := newBin(t)

	missing := t.TempDir()
	out, code := run(t, bin, missing, "{}", "hook", "session")
	if code != 0 {
		t.Fatalf("missing exit %d\n%s", code, out)
	}
	mustNotContain(t, out, sessionDirective)
	if strings.TrimSpace(out) != "" {
		t.Fatalf("missing roadmap should emit nothing, got:\n%s", out)
	}

	bad := t.TempDir()
	write(t, bad, ".workflow/roadmap.json", `{not valid`)
	out, code = run(t, bin, bad, "{}", "hook", "session")
	if code != 0 {
		t.Fatalf("invalid exit %d\n%s", code, out)
	}
	mustNotContain(t, out, sessionDirective)
}

// Half-A scenario 1: evidence JSON is not an active workflow; genuine one is.
func TestAcceptance_EvidenceJSONNotActive(t *testing.T) {
	bin := newBin(t)
	dir := t.TempDir()
	wfDir(t, dir)
	write(t, dir, ".workflow/alpha.json", `{"feature":"alpha","currentStep":"code","steps":{}}`)
	write(t, dir, ".workflow/alpha-qa-senior.json", `{"feature":"alpha","role":"qa-senior"}`)
	out, code := run(t, bin, dir, "{}", "hook", "context")
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	mustContain(t, out, "alpha")
	mustNotContain(t, out, "qa-senior")
}

// Half-A scenario 2: done workflow excluded while a genuine non-done one shows.
func TestAcceptance_DoneExcludedNonDoneShown(t *testing.T) {
	bin := newBin(t)
	dir := t.TempDir()
	doneWF(t, dir, "beta")
	write(t, dir, ".workflow/gamma.json", `{"feature":"gamma","currentStep":"tests","steps":{}}`)
	out, code := run(t, bin, dir, "{}", "hook", "context")
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	mustContain(t, out, "gamma")
	mustNotContain(t, out, "beta")
}

// Half-A scenario 3: ad-hoc roadmap JSONs are not treated as active workflows.
func TestAcceptance_AdHocRoadmapJSONsNotActive(t *testing.T) {
	bin := newBin(t)
	dir := t.TempDir()
	write(t, dir, ".workflow/roadmap.json", `{"phases":[]}`)
	write(t, dir, ".workflow/roadmap-quality.json", `{"role":"roadmap-quality-evaluator"}`)
	write(t, dir, ".workflow/delta.json", `{"feature":"delta","currentStep":"plan","steps":{}}`)
	out, code := run(t, bin, dir, "{}", "hook", "context")
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	mustContain(t, out, "delta")
	mustNotContain(t, out, "roadmap-quality")
}

// Half-A scenario 4: duplicate evidence JSONs collapse to a single panel row.
func TestAcceptance_DuplicatesDedupeToSingleRow(t *testing.T) {
	bin := newBin(t)
	dir := t.TempDir()
	write(t, dir, ".workflow/epsilon.json", `{"feature":"epsilon","currentStep":"code","steps":{}}`)
	write(t, dir, ".workflow/epsilon-big-thinker.json", `{"feature":"epsilon"}`)
	write(t, dir, ".workflow/epsilon-qa-senior.json", `{"feature":"epsilon"}`)
	out, code := run(t, bin, dir, "{}", "hook", "context")
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	// The dedupe contract is about the ACTIVE WORKFLOWS panel: epsilon must occupy
	// exactly one row there even though three epsilon-*.json files exist on disk.
	if n := strings.Count(activeWorkflowsPanel(out), "epsilon"); n != 1 {
		t.Fatalf("epsilon should appear exactly once in the active panel, saw %d:\n%s", n, out)
	}
}

// activeWorkflowsPanel returns just the rendered ACTIVE WORKFLOWS panel, trimmed
// before any subsequent panel (e.g. per-feature review-ready reminders).
func activeWorkflowsPanel(out string) string {
	start := strings.Index(out, "ACTIVE WORKFLOWS")
	if start < 0 {
		return out
	}
	panel := out[start:]
	if i := strings.Index(panel[1:], "🛡️👁️"); i >= 0 {
		return panel[:i+1]
	}
	return panel
}

// Half-A scenario 5: above the cap → 5 rows (most-recent) + "+2 more"; oldest hidden.
func TestAcceptance_CapShowsRecentPlusNMore(t *testing.T) {
	bin := newBin(t)
	dir := t.TempDir()
	wfDir(t, dir)
	base := time.Now().Add(-7 * time.Hour)
	for i := 0; i < 7; i++ {
		f := "feat-" + string(rune('0'+i))
		write(t, dir, ".workflow/"+f+".json",
			`{"feature":"`+f+`","currentStep":"code","steps":{}}`)
		mt := base.Add(time.Duration(i) * time.Hour)
		if err := os.Chtimes(filepath.Join(dir, ".workflow", f+".json"), mt, mt); err != nil {
			t.Fatal(err)
		}
	}
	out, code := run(t, bin, dir, "{}", "hook", "context")
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	// "+2 more" proves 7 active workflows were capped to 5 shown. The most-recent
	// feat-6 is among them. (Per-feature review-ready panels render every feature
	// name regardless of cap, so the precise 5-shown ordering is asserted in the
	// unit tier: ActiveWorkflows mtime-sort + CapActive front-N selection.)
	panel := activeWorkflowsPanel(out)
	mustContain(t, out, "+2 more")
	mustContain(t, panel, "feat-6") // most recent shown in the panel
	if c := countFeatures(panel); c != 5 {
		t.Fatalf("active panel should list exactly 5 features, listed %d:\n%s", c, panel)
	}
}

// countFeatures counts feat-N rows present in a rendered active-workflows panel.
func countFeatures(panel string) int {
	n := 0
	for i := 0; i < 7; i++ {
		if strings.Contains(panel, "feat-"+string(rune('0'+i))) {
			n++
		}
	}
	return n
}

// Half-A scenario 6: at-or-below the cap → no "+N more" hint.
func TestAcceptance_AtOrBelowCapNoMoreHint(t *testing.T) {
	bin := newBin(t)
	dir := t.TempDir()
	for _, f := range []string{"one", "two", "three"} {
		write(t, dir, ".workflow/"+f+".json",
			`{"feature":"`+f+`","currentStep":"code","steps":{}}`)
	}
	out, code := run(t, bin, dir, "{}", "hook", "context")
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	mustContain(t, out, "one")
	mustNotContain(t, out, "more active")
}
