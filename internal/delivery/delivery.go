// Package delivery is a read-only aggregator that composes the two delivery
// artifacts for `centinela deliver --via pr` — the Markdown PR body and a
// single Keep-a-Changelog line — out of the evidence Centinela already holds.
//
// It is PURE: callers (cmd/) read every source from disk, pass the file bodies
// in as plain strings, and this package returns rendered text. There is NO file
// I/O here (no os.ReadFile / os.WriteFile). Composition degrades section by
// section when a source is absent and NEVER fabricates a gate result. It may
// import internal/verify (read-only) for the VerificationResult type only.
package delivery

import "github.com/samuelnp/centinela/internal/verify"

// Evidence is the in-memory input the caller populates from disk. Every string
// is an already-read file body; an empty string means that source is missing,
// which makes the dependent section omit itself rather than error.
type Evidence struct {
	Feature          string // feature slug, e.g. "alpha"
	Brief            string // docs/features/<feature>.md body
	Plan             string // docs/plans/<feature>.md body
	GatekeeperReport string // .workflow/<feature>-gatekeeper.md body
	ChangelogStub    string // .workflow/<feature>-changelog.md body
	SpecPath         string // e.g. "specs/<feature>.feature" (empty = unknown)

	// Verification is an already-produced claim-verification result, or nil.
	// When nil the gate-status tally line is omitted (never faked).
	Verification *verify.VerificationResult
}

// ChangelogEntry is a rendered single changelog line and its target subsection.
// Category is exactly one of "Added", "Changed", or "Fixed" to match the
// existing Keep-a-Changelog subsections.
type ChangelogEntry struct {
	Category string
	Line     string
}
