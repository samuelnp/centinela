package verdict

import (
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/verify"
	"github.com/samuelnp/centinela/internal/workflow"
)

// Deps are the injected collaborators for one AssembleVerdict run. They keep
// the assembler pure: no time.Now, no os.Getwd, no real I/O wired in here.
type Deps struct {
	// Gates runs the gate suite (real: gates.RunAll — always a full scan in v1).
	Gates func(*config.Config) []gates.Result
	// Verify re-derives the feature's evidence claims for the current step.
	Verify func(feature, step string, cfg *config.Config) verify.VerificationResult
	// Evidence lists the on-disk role evidence for the feature.
	Evidence func(feature string) []EvidLine
	// Now is the injected RFC3339 timestamp (determinism).
	Now string
}

// AssembleVerdict aggregates gates + verify + evidence + a workflow/config
// snapshot into a deterministic packet. exitCode/verdict = fail iff any gate
// Fail OR verify HasFailures; warnings are reported but never fail.
func AssembleVerdict(feature string, cfg *config.Config, wf *workflow.Workflow, deps Deps) *Packet {
	gateResults := deps.Gates(cfg)
	step := ""
	if wf != nil {
		step = wf.CurrentStep
	}
	verifyResult := deps.Verify(feature, step, cfg)

	pkt := &Packet{
		Schema:   "centinela.verdict/v1",
		Run:      runInfo(feature, step, cfg, wf, deps.Now),
		Gates:    gateLines(gateResults),
		Verify:   checkLines(verifyResult.Checks),
		Evidence: deps.Evidence(feature),
	}
	pkt.Summary = summarize(gateResults, verifyResult)
	return pkt
}

func runInfo(feature, step string, cfg *config.Config, wf *workflow.Workflow, now string) RunInfo {
	archetype, _ := workflow.DisplayArchetype(wf)
	info := RunInfo{
		Feature:     feature,
		Step:        step,
		Profile:     workflow.EffectiveProfile(wf, cfg),
		Archetype:   archetype,
		Headless:    config.IsHeadless(cfg),
		GeneratedAt: now,
	}
	if wf != nil {
		info.DriverModel = wf.DriverModel
	}
	return info
}

func summarize(gateResults []gates.Result, verifyResult verify.VerificationResult) Summary {
	fail := !gates.AllPassed(gateResults) || verifyResult.HasFailures()
	s := Summary{
		Verdict:  "pass",
		ExitCode: 0,
		Gates:    gateCounts(gateResults),
		Verify:   verifyCounts(verifyResult),
	}
	if fail {
		s.Verdict = "fail"
		s.ExitCode = 1
	}
	return s
}

func gateLines(results []gates.Result) []GateLine {
	out := make([]GateLine, 0, len(results))
	for _, r := range results {
		out = append(out, gateLine(r))
	}
	return out
}

func checkLines(checks []verify.Check) []CheckLine {
	out := make([]CheckLine, 0, len(checks))
	for _, c := range checks {
		out = append(out, checkLine(c))
	}
	return out
}
