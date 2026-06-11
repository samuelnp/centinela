package gates

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gitdiff"
)

// checkSpecTraceability verifies that every in-scope Gherkin scenario maps to a
// covering acceptance test. Spec scope is diff-aware via filter (mirroring G1):
// a non-nil filter restricts parsing to changed .feature files. Coverage is
// global — a scenario may be covered by any acceptance file pointing at its
// spec, so the matcher always scans the whole test dir.
func checkSpecTraceability(cfg *config.Config, filter *gitdiff.Set) Result {
	c := cfg.Gates.SpecTraceability
	r := Result{Name: "spec-traceability-gate"}
	scenarios, err := parseScenarios(c.SpecDir, filter)
	if err != nil {
		r.Status = Fail
		r.Message = "spec-traceability: " + err.Error()
		return r
	}
	if len(scenarios) == 0 {
		r.Status = Skip
		r.Message = "No spec files in scope — gate skipped."
		return r
	}
	covered, err := coveredScenarios(c.TestDir)
	if err != nil {
		r.Status = Fail
		r.Message = "spec-traceability: " + err.Error()
		return r
	}
	return reportTraceability(scenarios, uncovered(scenarios, covered), c.Severity)
}

// reportTraceability maps the uncovered set to a Result: empty -> Pass with the
// covered count; otherwise Warn or Fail per severity, with one detail line per
// uncovered scenario naming its spec file and original scenario name.
func reportTraceability(scenarios, gaps []Scenario, severity string) Result {
	r := Result{Name: "spec-traceability-gate"}
	if len(gaps) == 0 {
		r.Status = Pass
		r.Message = fmt.Sprintf("All %d scenarios have acceptance coverage.", len(scenarios))
		return r
	}
	if severity == "warn" {
		r.Status = Warn
	} else {
		r.Status = Fail
	}
	r.Message = "Scenarios without acceptance coverage:"
	for _, s := range gaps {
		r.Details = append(r.Details, fmt.Sprintf("specs/%s.feature: %q", s.Spec, s.Name))
	}
	return r
}
