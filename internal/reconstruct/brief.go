package reconstruct

import "strings"

// briefStub assembles the docs/features/<slug>.md brief stub for a target as a
// pure, byte-stable string. It mirrors the docs/features/*.md shape (Problem /
// User value / What ships) with every unknown rendered as a "# TODO: confirm"
// marker, so the brief is honestly incomplete rather than fabricated.
func briefStub(t Target) string {
	var b strings.Builder
	b.WriteString("# Feature: " + t.Slug + "\n\n")
	b.WriteString("**Reconstructed from:** `" + t.Pkg + "` (" + string(roleOrModule(t.Role)) + ")\n")
	b.WriteString("**Selected because:** " + t.Reason + "\n")
	b.WriteString("**Status:** reconstructed-skeleton — confirm before adopting\n\n")
	b.WriteString("## Problem\n\n" + todoMarker + " — describe the problem this surface solves.\n\n")
	b.WriteString("## User value\n\n" + todoMarker + " — describe the value delivered to the user.\n\n")
	b.WriteString("## What ships\n\n- " + todoMarker + " — enumerate the behaviors this surface ships.\n\n")
	b.WriteString("## Acceptance criteria\n\n")
	b.WriteString("- " + todoMarker + " — see `specs/" + t.Slug + ".feature` and replace each placeholder step.\n")
	return b.String()
}
