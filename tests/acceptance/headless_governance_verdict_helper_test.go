package acceptance_test

import (
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/verdict"
	"github.com/samuelnp/centinela/internal/verify"
	"github.com/samuelnp/centinela/internal/workflow"
)

const hgNow = "2026-06-12T00:00:00Z"

// hgDeps fully injects gates/verify/evidence + a fixed Now so AssembleVerdict
// runs golden/byte-stable without touching real gates, verify, or the disk.
func hgDeps(g []gates.Result, v verify.VerificationResult, e []verdict.EvidLine) verdict.Deps {
	return verdict.Deps{
		Gates:    func(*config.Config) []gates.Result { return g },
		Verify:   func(string, string, *config.Config) verify.VerificationResult { return v },
		Evidence: func(string) []verdict.EvidLine { return e },
		Now:      hgNow,
	}
}

func hgPassGate() gates.Result {
	return gates.Result{Name: "G1: File Size", Status: gates.Pass, Message: "ok"}
}

func hgVerify(checks ...verify.Check) verify.VerificationResult {
	return verify.VerificationResult{Feature: "headless-governance", Checks: checks}
}

func hgWf() *workflow.Workflow {
	return &workflow.Workflow{Feature: "headless-governance", CurrentStep: "validate", DriverModel: "claude-opus"}
}
