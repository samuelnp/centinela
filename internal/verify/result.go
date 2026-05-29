// Package verify independently re-derives ground truth for the claims an
// evidence file makes and reports divergence. It is a domain package: it
// imports config/evidence/orchestration/worktree but never cmd/ or ui/.
package verify

// Status is the outcome of one claim check.
type Status string

const (
	// StatusPass — the claim was independently confirmed.
	StatusPass Status = "PASS"
	// StatusFail — the claim diverges from ground truth (hard block).
	StatusFail Status = "FAIL"
	// StatusSkip — no such claim was made; nothing to verify.
	StatusSkip Status = "SKIP"
	// StatusWarn — divergence on a heuristic check; reported, not blocking.
	StatusWarn Status = "WARN"
	// StatusConfigError — the check could not run due to misconfiguration.
	StatusConfigError Status = "CONFIG-ERROR"
	// StatusTimeout — the check exceeded the configured timeout.
	StatusTimeout Status = "TIMEOUT"
)

// Check is a single claim verified against ground truth.
type Check struct {
	Claim  string
	Role   string
	Status Status
	Detail string
}

// VerificationResult aggregates every claim check for a feature.
type VerificationResult struct {
	Feature string
	Checks  []Check
}

// blocking reports whether a status hard-blocks completion. WARN and SKIP do
// not; everything that leaves a claim unconfirmed does.
func (s Status) blocking() bool {
	switch s {
	case StatusFail, StatusConfigError, StatusTimeout:
		return true
	}
	return false
}

// HasFailures reports whether any check leaves a claim unconfirmed (FAIL,
// CONFIG-ERROR, or TIMEOUT). WARN and SKIP do not count.
func (r VerificationResult) HasFailures() bool {
	for _, c := range r.Checks {
		if c.Status.blocking() {
			return true
		}
	}
	return false
}

// HasWarnings reports whether any check produced a non-blocking warning.
func (r VerificationResult) HasWarnings() bool {
	for _, c := range r.Checks {
		if c.Status == StatusWarn {
			return true
		}
	}
	return false
}

// Failed returns only the blocking checks, for naming in error messages.
func (r VerificationResult) Failed() []Check {
	var out []Check
	for _, c := range r.Checks {
		if c.Status.blocking() {
			out = append(out, c)
		}
	}
	return out
}

// Tally counts checks by broad outcome for the summary line.
func (r VerificationResult) Tally() (pass, fail, skip, warn int) {
	for _, c := range r.Checks {
		switch c.Status {
		case StatusPass:
			pass++
		case StatusSkip:
			skip++
		case StatusWarn:
			warn++
		default:
			fail++
		}
	}
	return pass, fail, skip, warn
}
