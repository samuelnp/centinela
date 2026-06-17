package verdict

import (
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/verify"
)

// gateStatus lowercases a gate Status into the packet vocabulary.
func gateStatus(s gates.Status) string {
	switch s {
	case gates.Pass:
		return "pass"
	case gates.Fail:
		return "fail"
	case gates.Warn:
		return "warn"
	case gates.Skip:
		return "skip"
	}
	return "unknown"
}

// gateLine maps one gate Result to a GateLine.
func gateLine(r gates.Result) GateLine {
	return GateLine{
		Name:    r.Name,
		Status:  gateStatus(r.Status),
		Message: r.Message,
		Details: r.Details,
	}
}

// checkLine maps one verify Check to a CheckLine (status stays UPPERCASE).
func checkLine(c verify.Check) CheckLine {
	return CheckLine{
		Role:   c.Role,
		Claim:  c.Claim,
		Status: string(c.Status),
		Detail: c.Detail,
	}
}

// gateCounts tallies gate results by lowercased status.
func gateCounts(results []gates.Result) Counts {
	var c Counts
	for _, r := range results {
		switch r.Status {
		case gates.Pass:
			c.Pass++
		case gates.Fail:
			c.Fail++
		case gates.Warn:
			c.Warn++
		case gates.Skip:
			c.Skip++
		}
	}
	return c
}

// verifyCounts tallies a VerificationResult via its native Tally.
func verifyCounts(res verify.VerificationResult) Counts {
	pass, fail, skip, warn := res.Tally()
	return Counts{Pass: pass, Fail: fail, Warn: warn, Skip: skip}
}
