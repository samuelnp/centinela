package verify

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/evidence"
)

const (
	claimCoverage = "coverage-moved"
	coverageCmd   = "go test -cover ./..."
)

// coverageLine matches `coverage: 84.9% of statements` in `go test` output.
var coverageLine = regexp.MustCompile(`coverage:\s+([0-9.]+)% of statements`)

// checkCoverage re-derives per-package coverage (no -coverpkg, matching the
// project model) and fails when the claimed figure sits above the measured
// figure by more than the configured tolerance. Absent claim → SKIP.
func checkCoverage(cfg *config.Config, deps Deps, role string, ev *evidence.RoleEvidence, timeout time.Duration) Check {
	c := Check{Claim: claimCoverage, Role: role}
	if ev.Coverage == nil {
		c.Status = StatusSkip
		c.Detail = "no coverage claim in evidence"
		return c
	}
	claimed := *ev.Coverage
	out := deps.Runner.Run(deps.Root, coverageCmd, timeout)
	switch {
	case out.TimedOut:
		c.Status = StatusTimeout
		c.Detail = fmt.Sprintf("coverage command %q exceeded the verify timeout", coverageCmd)
		return c
	case out.StartErr != nil:
		c.Status = StatusConfigError
		c.Detail = fmt.Sprintf("coverage command %q could not start: %s", coverageCmd, out.StartErr)
		return c
	}
	measured, ok := meanCoverage(out.Output)
	if !ok {
		c.Status = StatusConfigError
		c.Detail = "no per-package coverage figures found in `go test -cover` output"
		return c
	}
	return classifyCoverage(c, claimed, measured, cfg.Verify.CoverageTolerance)
}

// classifyCoverage applies the claim-above-measured-beyond-tolerance rule.
// Tolerance is fractional (0.001 = 0.1 percentage points).
func classifyCoverage(c Check, claimed, measured, tolerance float64) Check {
	tolPct := tolerance * 100
	if claimed-measured > tolPct {
		c.Status = StatusFail
		c.Detail = fmt.Sprintf("claimed %.2f%% but measured %.2f%% (gap exceeds %.3f%% tolerance)",
			claimed, measured, tolPct)
		return c
	}
	c.Status = StatusPass
	c.Detail = fmt.Sprintf("claimed %.2f%%, measured %.2f%% (within %.3f%% tolerance)", claimed, measured, tolPct)
	return c
}

// meanCoverage averages every per-package coverage figure in out. ok is false
// when no figures are present.
func meanCoverage(out string) (float64, bool) {
	matches := coverageLine.FindAllStringSubmatch(out, -1)
	if len(matches) == 0 {
		return 0, false
	}
	var sum float64
	for _, m := range matches {
		v, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			return 0, false
		}
		sum += v
	}
	return sum / float64(len(matches)), true
}
