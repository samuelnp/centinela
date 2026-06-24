package audit

import (
	"github.com/samuelnp/centinela/internal/config"
)

// Outcome is the result of an Adopt call. It carries the skip-if-exists verdict,
// the configured baseline path, and the recorded baseline (zero value when the
// adoption was skipped because a baseline already existed and --force was off).
type Outcome struct {
	Skipped  bool     // true when a baseline already existed and !force (no write performed)
	Path     string   // configured baseline path
	Baseline Baseline // the recorded baseline (zero value when Skipped)
}

// Adopt is the one-time brownfield adoption orchestrator. It owns the
// skip-if-exists business rule (kept out of cmd/ per G7): if a baseline already
// exists at the configured path and force is false, it returns a Skipped outcome
// WITHOUT writing anything; otherwise it records the current full-repo violations
// and Saves them, returning the recorded baseline for the report. The written
// file is byte-identical to Record+Save — adopt adds semantics, not data.
func Adopt(cfg *config.Config, force bool) (Outcome, error) {
	path := cfg.Gates.AuditBaseline.BaselinePath
	_, exists, err := Load(path)
	if err != nil {
		return Outcome{}, err
	}
	if exists && !force {
		return Outcome{Skipped: true, Path: path}, nil
	}
	b := Record(cfg)
	if err := Save(path, b); err != nil {
		return Outcome{}, err
	}
	return Outcome{Skipped: false, Path: path, Baseline: b}, nil
}

// Total reports the number of baselined fingerprints across all gates.
func (b Baseline) Total() int {
	n := 0
	for _, e := range b.Gates {
		n += len(e.Fingerprints)
	}
	return n
}
