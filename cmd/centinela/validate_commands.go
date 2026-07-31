package main

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/acceptance"
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/ui"
)

// runValidateCommands executes every configured validate command and reports a
// three-valued verdict per command. Only a Fail flips the run, so a Warn
// (an undetermined acceptance report) is visible without being load-bearing.
func runValidateCommands(cfg *config.Config) bool {
	lastValidateRun = validateRunRecord{}
	if len(cfg.Validate.Commands) == 0 {
		return true
	}
	fmt.Println()
	fmt.Println(ui.StyleBold.Render("Validate Commands"))
	allPassed := true
	for _, cmd := range cfg.Validate.Commands {
		passed, out := runCommand(cmd)
		status, detail := commandVerdict(cfg, cmd, passed, out)
		lastValidateRun.record(cmd, out, passed, status, detail, cfg.Validate.AcceptanceSkipPolicy)
		fmt.Println(ui.RenderCmdVerdict(cmd, status, detail, out))
		if status == gates.Fail {
			allPassed = false
		}
	}
	return allPassed
}

// commandVerdict maps one command outcome onto the gate status vocabulary.
//
// Order is the contract: the EXIT CODE wins, then the classifier gates the
// parse, then the parse gates the verdict. A command that both fails and
// reports skips is reported as the exit failure; a non-acceptance command that
// legitimately skips (build tags, -short, platform gates) is never touched.
//
// The invariant holds per TIER, not merely per command: acceptance.Scope makes
// a whole-repo command count only the results it can attribute to the
// acceptance path, so a unit tier's skip cannot fail an acceptance gate even
// when one combined command runs everything.
func commandVerdict(cfg *config.Config, cmd string, passed bool, out string) (gates.Status, string) {
	if !passed {
		return gates.Fail, ""
	}
	if !acceptance.IsExecutionCommand(cmd) {
		return gates.Pass, ""
	}
	verdict, detail := acceptance.Judge(cmd, out, cfg.Validate.AcceptanceSkipPolicy)
	return acceptanceStatus(verdict), detail
}

// acceptanceStatus translates the leaf's verdict vocabulary into gate statuses.
func acceptanceStatus(v acceptance.Verdict) gates.Status {
	switch v {
	case acceptance.VerdictFail:
		return gates.Fail
	case acceptance.VerdictWarn:
		return gates.Warn
	default:
		return gates.Pass
	}
}
