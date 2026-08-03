package gates

import (
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gitdiff"
)

// Status represents the outcome of a gate check.
type Status int

// The gate outcomes. Fail is the only status that blocks a run; Skip means the
// gate inspected nothing and therefore refuses to claim a pass.
const (
	// Pass means the gate inspected its scope and found no violation.
	Pass Status = iota
	// Fail means the gate found a blocking violation.
	Fail
	// Warn means the gate found a violation it was configured not to block on.
	Warn
	// Skip means the gate had nothing in scope to inspect.
	Skip
)

// Result is the outcome of a single gate.
type Result struct {
	Name    string
	Status  Status
	Message string
	Details []string
}

// RunAll executes all enabled built-in gates in full-scan mode.
// Kept for backward compatibility; equivalent to RunWithFilter(cfg, nil).
func RunAll(cfg *config.Config) []Result {
	return RunWithFilter(cfg, nil)
}

// RunWithFilter executes all enabled built-in gates. When filter is non-nil
// the file-scoped gates (G1, G11) only inspect files in the filter set.
// A nil filter preserves the legacy whole-repo behavior.
func RunWithFilter(cfg *config.Config, filter *gitdiff.Set) []Result {
	var results []Result

	if cfg.Gates.FileSizeEnabled {
		results = append(results, checkFileSize(cfg, filter))
	}

	if cfg.Gates.I18nEnabled {
		results = append(results, checkI18nFiltered(cfg, filter))
	}

	if cfg.Gates.Build.Enabled {
		results = append(results, checkBuild(cfg))
	}

	if cfg.Gates.ImportGraph.Enabled {
		results = append(results, checkImportGraph(cfg, filter))
	}

	if cfg.Gates.Security.Enabled {
		results = append(results, checkSecurity(cfg, filter)...)
	}

	if cfg.Gates.SpecTraceability.Enabled {
		results = append(results, checkSpecTraceability(cfg, filter))
	}

	if cfg.Gates.RoadmapDrift.Enabled {
		results = append(results, checkRoadmapDrift(cfg, filter))
	}

	if cfg.Gates.Docstring.Enabled {
		results = append(results, checkDocstring(cfg, filter))
	}

	if len(cfg.Gates.CustomGates) > 0 {
		results = append(results, customGates(cfg, filter)...)
	}

	return results
}

// AllPassed returns true if no result has a Fail status.
func AllPassed(results []Result) bool {
	for _, r := range results {
		if r.Status == Fail {
			return false
		}
	}
	return true
}
