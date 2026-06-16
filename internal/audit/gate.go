package audit

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
)

// gateName is the Result.Name of the optional audit-baseline gate wired into
// validate from cmd/ (gates may not import audit — see plan cycle-avoidance note).
const gateName = "audit_baseline"

// Check is the audit_baseline gate body. A missing baseline is non-blocking
// (Skip) so a repo that hasn't opted in never fails. Otherwise it ratchets and
// maps any new violation to Fail or Warn per the configured severity. Details
// are the new fingerprints' Raw lines for human review.
func Check(cfg *config.Config) gates.Result {
	path := cfg.Gates.AuditBaseline.BaselinePath
	b, exists, err := Load(path)
	if err != nil {
		return gates.Result{Name: gateName, Status: gates.Fail, Message: err.Error()}
	}
	if !exists {
		return gates.Result{
			Name:    gateName,
			Status:  gates.Skip,
			Message: "no baseline; run `centinela audit baseline`",
		}
	}
	if b.SchemeStale() {
		return gates.Result{
			Name:    gateName,
			Status:  gates.Warn,
			Message: "baseline scheme changed — re-run `centinela audit baseline`",
		}
	}

	d := Ratchet(cfg, b)
	if !d.HasNew() {
		return gates.Result{
			Name:    gateName,
			Status:  gates.Pass,
			Message: fmt.Sprintf("no new violations (%d baselined)", len(d.Baselined)),
		}
	}
	return gates.Result{
		Name:    gateName,
		Status:  newStatus(cfg.Gates.AuditBaseline.Severity),
		Message: fmt.Sprintf("%d new violation(s) since baseline", len(d.New)),
		Details: rawLines(d.New),
	}
}

// newStatus maps the configured severity to the gate Status for new violations.
func newStatus(severity string) gates.Status {
	if severity == "warn" {
		return gates.Warn
	}
	return gates.Fail
}

func rawLines(fps []Fingerprint) []string {
	out := make([]string, 0, len(fps))
	for _, fp := range fps {
		out = append(out, fp.Raw)
	}
	return out
}
