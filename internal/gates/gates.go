package gates

import "github.com/samuelnp/centinela/internal/config"

// Status represents the outcome of a gate check.
type Status int

const (
	Pass Status = iota
	Fail
	Warn
	Skip
)

// Result is the outcome of a single gate.
type Result struct {
	Name    string
	Status  Status
	Message string
	Details []string
}

// RunAll executes all enabled built-in gates and returns their results.
func RunAll(cfg *config.Config) []Result {
	var results []Result

	if cfg.Gates.FileSizeEnabled {
		results = append(results, checkFileSize())
	}

	if cfg.Gates.I18nEnabled {
		results = append(results, checkI18n(cfg))
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
