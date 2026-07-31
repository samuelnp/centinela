package config

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

// DocstringConfig controls the docstring gate. When Enabled, the gate reports
// every exported identifier in a *changed* file that carries no doc comment.
// Scope is always the changed-file set — there is deliberately no scope knob
// here, because scope is a repo-wide decision made by [validate] diff_mode /
// diff_base. Severity ("fail"|"warn") is the adoption knob; this project ships
// "fail", since the ratchet already keeps a legacy backlog out of scope.
type DocstringConfig struct {
	// Enabled turns the gate on. Absent block means off.
	Enabled bool `toml:"enabled"`
	// Severity is "fail" (default) or "warn".
	Severity string `toml:"severity"`
	// Roots limits the scan to files under these directories.
	Roots []string `toml:"roots"`
	// IncludeInternal is a tri-state knob; unset means true.
	IncludeInternal *bool `toml:"include_internal"`
	// CheckFields extends the check to exported struct fields.
	CheckFields bool `toml:"check_fields"`
	// RequireNamePrefix demands the doc comment open with the identifier.
	RequireNamePrefix bool `toml:"require_name_prefix"`
}

// IncludesInternal resolves the tri-state include_internal knob. Unset means
// true: this repo is almost entirely internal/, and excluding it by default
// would make the gate vacuous here.
func (d DocstringConfig) IncludesInternal() bool {
	return d.IncludeInternal == nil || *d.IncludeInternal
}

// NormalizeDocstring trims whitespace, cleans and drops blank roots, and fills
// the unset severity and include_internal defaults (fail, true).
//
// Cleaning the roots is a correctness fix, not tidiness: scanned paths arrive
// repo-relative from git, so an unnormalized "./internal" matched nothing and
// an enforcing gate silently reported "no Go files in scope" on a tree it
// should have failed. A config spelling must never disarm a gate.
func NormalizeDocstring(d DocstringConfig) DocstringConfig {
	d.Severity = strings.TrimSpace(d.Severity)
	if d.Severity == "" {
		d.Severity = "fail"
	}
	roots := make([]string, 0, len(d.Roots))
	for _, r := range d.Roots {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		if !filepath.IsAbs(r) {
			r = path.Clean(filepath.ToSlash(r))
		}
		roots = append(roots, r)
	}
	d.Roots = roots
	if d.IncludeInternal == nil {
		on := true
		d.IncludeInternal = &on
	}
	return d
}

// validateDocstring rejects an unknown severity so a typo cannot silently
// relax the gate. It is a no-op when the gate is disabled.
func validateDocstring(d DocstringConfig) error {
	if !d.Enabled {
		return nil
	}
	if d.Severity != "fail" && d.Severity != "warn" {
		return fmt.Errorf("gates.docstring.severity must be fail or warn, got %q", d.Severity)
	}
	for i, r := range d.Roots {
		// Scanned paths are repo-relative, so an absolute root can never match.
		// Failing closed beats scanning nothing and calling it a pass.
		if filepath.IsAbs(r) {
			return fmt.Errorf(
				"gates.docstring.roots[%d] must be repo-relative, got absolute path %q", i, r)
		}
	}
	return nil
}
