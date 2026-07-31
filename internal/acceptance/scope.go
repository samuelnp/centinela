package acceptance

import "strings"

// AcceptancePath is the repository path that identifies the acceptance tier.
// It is the same literal the classifier keys on, so classification and
// attribution can never disagree about where acceptance tests live.
const AcceptancePath = "tests/acceptance"

// Scope answers a question the classifier alone cannot: "is every result this
// command reports acceptance work?" That is NOT the same as "does this command
// execute acceptance tests". A whole-repo `go test -v ./...` answers yes to the
// second and no to the first — its output also carries the unit tier's
// build-tag, -short and platform skips. Treating those as acceptance skips is
// an over-block, and it breaks AC5 at the TIER level even while holding at the
// command level.
type Scope int

const (
	// ScopeNone — not an acceptance execution. Never analysed at all.
	ScopeNone Scope = iota
	// ScopeAcceptance — the command itself restricts execution to the
	// acceptance tier (a path-scoped run, or a dedicated Gherkin runner), so
	// every result it reports is acceptance work.
	ScopeAcceptance
	// ScopeMixed — a whole-repo command. Go results count only when
	// attributable to a package under AcceptancePath; anything that cannot be
	// attributed is never failed.
	ScopeMixed
)

// ScopeOf classifies ONE command string. The union of the non-None branches is
// byte-identical to the pre-move acceptance predicate — this splits the
// predicate's answer, it does not change which commands it accepts.
func ScopeOf(cmd string) Scope {
	c := strings.ToLower(strings.TrimSpace(cmd))
	if c == "" {
		return ScopeNone
	}
	switch {
	case strings.Contains(c, AcceptancePath),
		strings.Contains(c, "cucumber"),
		strings.Contains(c, "godog"),
		strings.Contains(c, "behave"),
		strings.Contains(c, "acceptance") && strings.Contains(c, "test"):
		return ScopeAcceptance
	case strings.Contains(c, "go test") && strings.Contains(c, "./..."):
		return ScopeMixed
	}
	return ScopeNone
}

// counts reports whether a Go result produced by pkg is in the tier under
// analysis. Under ScopeMixed an unattributable result (empty package) is
// deliberately NOT counted: blaming the acceptance tier for a skip we cannot
// place is the over-block this scope exists to prevent.
func (s Scope) counts(pkg string) bool {
	if s != ScopeMixed {
		return true
	}
	return isAcceptancePackage(pkg)
}

// isAcceptancePackage matches a go-test package path (e.g.
// "github.com/x/y/tests/acceptance") against the acceptance tier.
func isAcceptancePackage(pkg string) bool {
	p := strings.ReplaceAll(strings.TrimSpace(pkg), "\\", "/")
	if p == "" {
		return false
	}
	return p == AcceptancePath ||
		strings.HasPrefix(p, AcceptancePath+"/") ||
		strings.HasSuffix(p, "/"+AcceptancePath) ||
		strings.Contains(p, "/"+AcceptancePath+"/")
}
