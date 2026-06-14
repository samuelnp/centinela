// Package doctor diagnoses (and, where safe, repairs) Centinela project-health
// problems: hook wiring, roadmap drift, abandoned worktrees, stale .workflow
// state, orphaned evidence, config drift, and binary version skew. Every check
// is pure (Run never mutates); repairs run only under Fix and only when Safe.
package doctor

import "github.com/samuelnp/centinela/internal/config"

// Status is the severity of a single check's diagnosis.
type Status int

const (
	// OK means the check found no problem.
	OK Status = iota
	// Warn means an advisory problem that does not fail the command.
	Warn
	// Error means a problem that fails the command (exit 1).
	Error
)

// Repair describes how a diagnosis can be remediated. Safe+Idempotent repairs
// are eligible for --fix and carry a non-nil Apply. Report-only/destructive
// remediations leave Apply nil and set Command to the exact user-runnable line.
type Repair struct {
	Safe       bool         // true => eligible for --fix
	Idempotent bool         // documents the re-run guarantee
	Apply      func() error // nil for report-only/destructive checks
	Command    string       // user-runnable command for report-only fixes
}

// Diagnosis is the result of running one check. Name is a stable identifier
// that drives both ordering and rendering. Details holds supplementary lines.
type Diagnosis struct {
	Name    string
	Status  Status
	Message string
	Details []string
	Repair  *Repair // nil when nothing to fix
}

// Check diagnoses one aspect of project health. Run must be pure: it inspects
// the repo via ctx and returns a Diagnosis without mutating any file.
type Check interface {
	Name() string
	Run(ctx Context) Diagnosis
}

// Context carries the resolved repo root and the loaded config so each check is
// pure and unit-testable against a temp dir. Checks run with the process CWD
// already set to Root (see context.go), so they may reuse CWD-relative domain
// APIs (config.Load, roadmap.Load, setup, evidence.Repair) unchanged.
type Context struct {
	Root   string         // canonical repo root (never the worktree subtree)
	Config *config.Config // nil when centinela.toml failed to parse
	CfgErr error          // non-nil when centinela.toml could not be parsed
}
