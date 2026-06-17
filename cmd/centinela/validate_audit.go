package main

import (
	"github.com/samuelnp/centinela/internal/audit"
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
)

// appendAuditGate wires the audit_baseline gate into the validate result set
// from cmd/, not inside gates.RunWithFilter, because gates (domain) may not
// import audit (aggregator) — that would create a cycle and break the
// import-graph matrix. cmd is the correct seam. No-op unless the gate is enabled.
func appendAuditGate(cfg *config.Config, results []gates.Result) []gates.Result {
	if cfg.Gates.AuditBaseline.Enabled {
		results = append(results, audit.Check(cfg))
	}
	return results
}
