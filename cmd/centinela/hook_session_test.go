package main

import (
	"strings"
	"testing"
)

const sessionDirective = "CENTINELA DIRECTIVE: session rehydration"

// runSession drives the real SessionStart hook end-to-end and captures stdout.
func runSession(t *testing.T) string {
	t.Helper()
	var out string
	withStdin(t, `{"source":"startup"}`, func() {
		out = captureStdout(t, func() {
			if err := runHookSession(nil, nil); err != nil {
				t.Fatalf("runHookSession returned error: %v", err)
			}
		})
	})
	return out
}

// Spec Half-B: valid roadmap → directive + roadmap + next feature + both pointers.
func TestRunHookSession_ValidRoadmapEmitsRehydration(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, "PROJECT.md", "x")
	writeFile(t, ".workflow/roadmap.json",
		`{"phases":[{"name":"Phase 0","features":[{"name":"next-feature"},{"name":"later"}]}]}`)
	out := runSession(t)
	for _, want := range []string{
		sessionDirective, "next-feature", "(planned)",
		"PROJECT.md", "docs/features/next-feature.md",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected session output to contain %q, got:\n%s", want, out)
		}
	}
}

// Spec: next feature is first incomplete across ALL phases (not just Phase 0).
func TestRunHookSession_NextIsFirstIncompleteAcrossPhases(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, "PROJECT.md", "x")
	writeFile(t, ".workflow/roadmap.json",
		`{"phases":[{"name":"Phase 0","features":[{"name":"p0a"},{"name":"p0b"}]},`+
			`{"name":"Phase 1","features":[{"name":"phase-1-first"}]}]}`)
	markCmdDone(t, "p0a")
	markCmdDone(t, "p0b")
	out := runSession(t)
	if !strings.Contains(out, "Next feature to plan: phase-1-first") {
		t.Fatalf("expected phase-1-first as next, got:\n%s", out)
	}
	if !strings.Contains(out, "docs/features/phase-1-first.md") {
		t.Fatalf("expected phase-1-first pointer, got:\n%s", out)
	}
}

// Spec: every feature done → roadmap-complete, no <next>.md pointer, exit zero.
func TestRunHookSession_AllDoneRoadmapComplete(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, "PROJECT.md", "x")
	writeFile(t, ".workflow/roadmap.json",
		`{"phases":[{"name":"P","features":[{"name":"only"}]}]}`)
	markCmdDone(t, "only")
	out := runSession(t)
	if !strings.Contains(out, "Roadmap complete") {
		t.Fatalf("expected roadmap-complete line, got:\n%s", out)
	}
	if strings.Contains(out, "docs/features/") {
		t.Fatalf("complete state must not emit a <next>.md pointer, got:\n%s", out)
	}
}

// Spec: missing roadmap and invalid roadmap json both exit zero with no payload.
func TestRunHookSession_MissingAndInvalidAreSilent(t *testing.T) {
	chdirIntoTemp(t)
	if out := runSession(t); strings.TrimSpace(out) != "" {
		t.Fatalf("missing roadmap should emit nothing, got:\n%s", out)
	}
	writeFile(t, ".workflow/roadmap.json", `{not valid json`)
	out := runSession(t)
	if strings.Contains(out, sessionDirective) || strings.TrimSpace(out) != "" {
		t.Fatalf("invalid roadmap should emit nothing, got:\n%s", out)
	}
}
