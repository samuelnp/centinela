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

// Spec Half-B: valid roadmap → directive + roadmap + plural ready frontier + pointers.
// No-dep planned features derive to "ready", so the body annotates them (ready).
func TestRunHookSession_ValidRoadmapEmitsRehydration(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, "PROJECT.md", "x")
	writeFile(t, ".workflow/roadmap.json",
		`{"phases":[{"name":"Phase 0","features":[{"name":"next-feature"},{"name":"later"}]}]}`)
	out := runSession(t)
	for _, want := range []string{
		sessionDirective, "next-feature", "(ready)",
		"Ready to start now:",
		"PROJECT.md", "docs/features/next-feature.md",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected session output to contain %q, got:\n%s", want, out)
		}
	}
}

// Spec: rehydration lists the full ready frontier across ALL phases (not just
// Phase 0). With p0a/p0b done, the no-dep Phase 1 feature is ready.
func TestRunHookSession_ListsReadyFrontier(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, "PROJECT.md", "x")
	writeFile(t, ".workflow/roadmap.json",
		`{"phases":[{"name":"Phase 0","features":[{"name":"p0a"},{"name":"p0b"}]},`+
			`{"name":"Phase 1","features":[{"name":"phase-1-first"}]}]}`)
	markCmdDone(t, "p0a")
	markCmdDone(t, "p0b")
	out := runSession(t)
	if !strings.Contains(out, "Ready to start now:") {
		t.Fatalf("expected ready frontier header, got:\n%s", out)
	}
	if !strings.Contains(out, "phase-1-first") {
		t.Fatalf("expected phase-1-first in ready frontier, got:\n%s", out)
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
