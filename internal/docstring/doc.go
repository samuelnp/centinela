// Package docstring is a leaf that reports exported identifiers declared
// without a doc comment. It walks the syntax tree of an explicit list of files
// and returns a Report which the docstring gate and `centinela docs lint`
// render — it never resolves scope itself, so there is exactly one notion of
// "changed" in the codebase and it lives in internal/gitdiff.
//
// The package depends on the standard library only. That is deliberate:
// internal/gates (domain) may import a leaf, and a stdlib-only package cannot
// participate in an import cycle. See PROJECT.md G2.
package docstring

// Directive is the opt-out marker. It uses Go's directive form (no space after
// the slashes) and Centinela's own namespace — deliberately not //nolint,
// which belongs to golangci-lint and would couple our semantics to theirs.
const Directive = "//centinela:nodoc"

// DefaultRoots are the source roots inspected when Options.Roots is unset.
var DefaultRoots = []string{"internal", "cmd", "pkg", "lib", "src", "app"}

// Options tunes a scan. The zero value inspects DefaultRoots and excludes
// internal packages, struct fields and name-prefix strictness.
type Options struct {
	// Roots limits the scan to files under these directories.
	Roots []string
	// IncludeInternal keeps files under an internal/ directory in scope.
	IncludeInternal bool
	// CheckFields extends the check to exported struct fields.
	CheckFields bool
	// RequireNamePrefix demands the doc comment start with the identifier.
	RequireNamePrefix bool
}

// roots returns the effective root list for opts.
func roots(opts Options) []string {
	if len(opts.Roots) == 0 {
		return DefaultRoots
	}
	return opts.Roots
}
