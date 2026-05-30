package verify

import (
	"time"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/evidence"
	"github.com/samuelnp/centinela/internal/orchestration"
)

// EvidenceLoader fetches the on-disk evidence for a (feature, role) pair.
// Injected so verification is testable without touching the filesystem.
type EvidenceLoader func(feature string, role orchestration.Role) (*evidence.RoleEvidence, error)

// Deps are the injected collaborators for one Verify run.
type Deps struct {
	// Root is the directory verification resolves paths against (the active
	// worktree root, or the repo root when worktrees are off).
	Root string
	// Runner executes test/coverage commands.
	Runner CommandRunner
	// Load reads evidence; defaults to evidence.Read when nil.
	Load EvidenceLoader
	// PriorTestRun, when non-nil, is reused by the tests-pass check instead of
	// re-running the suite (the complete gate already ran it once).
	PriorTestRun *RunOutcome
}

// Verify re-derives ground truth for the claims in the feature's evidence for
// the given workflow step and returns a per-claim result. It never mutates
// state; it only reads evidence and runs read-only commands.
func Verify(feature, step string, cfg *config.Config, deps Deps) VerificationResult {
	if deps.Load == nil {
		deps.Load = evidence.Read
	}
	res := VerificationResult{Feature: feature}
	roles := orchestration.RequiredRoles(step)
	for _, role := range roles {
		ev, err := deps.Load(feature, role)
		if err != nil || ev == nil {
			continue // no evidence for this role — nothing to verify
		}
		res.Checks = append(res.Checks, runChecks(cfg, deps, role, ev)...)
	}
	if len(res.Checks) == 0 {
		res.Checks = append(res.Checks, Check{
			Claim:  "claims",
			Status: StatusSkip,
			Detail: "no claims to verify",
		})
	}
	return res
}

// runChecks runs the four claim checks for one role's evidence.
func runChecks(cfg *config.Config, deps Deps, role orchestration.Role, ev *evidence.RoleEvidence) []Check {
	timeout := time.Duration(cfg.Verify.TimeoutSeconds) * time.Second
	return []Check{
		checkTestsPass(cfg, deps, string(role), timeout),
		checkCoverage(cfg, deps, string(role), ev, timeout),
		checkStubs(deps.Root, string(role), ev),
		checkEdgeCases(deps.Root, string(role), ev),
	}
}
