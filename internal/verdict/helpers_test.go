package verdict

import (
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/verify"
)

const fixedNow = "2026-06-12T00:00:00Z"

// fakeDeps builds a fully-injected Deps so AssembleVerdict runs with no real
// I/O: gates/verify/evidence are constant closures and Now is fixed.
func fakeDeps(g []gates.Result, v verify.VerificationResult, e []EvidLine) Deps {
	return Deps{
		Gates:    func(*config.Config) []gates.Result { return g },
		Verify:   func(string, string, *config.Config) verify.VerificationResult { return v },
		Evidence: func(string) []EvidLine { return e },
		Now:      fixedNow,
	}
}

func passGate() gates.Result {
	return gates.Result{Name: "G1: File Size", Status: gates.Pass, Message: "ok"}
}

func failGate() gates.Result {
	return gates.Result{Name: "G1: File Size", Status: gates.Fail, Message: "too long"}
}

func vr(checks ...verify.Check) verify.VerificationResult {
	return verify.VerificationResult{Feature: "headless-governance", Checks: checks}
}
