package main

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/acceptance"
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/verify"
)

// validateRunRecord is this process's own record of the validate commands it
// just ran: the real captured output, plus the acceptance-skip judgement the
// per-command analysis ALREADY produced — each command judged with that
// command's own scope.
//
// It exists because the reused single-run outcome stands for several commands
// at once. Re-deriving a verdict from their concatenated transcript could
// neither tell which command a line came from nor recover each command's scope,
// so `complete` inherits this record instead of re-parsing prose.
type validateRunRecord struct {
	outputs  []string
	notes    []string
	analysed bool
	failed   bool
	disabled bool
}

// lastValidateRun is reset at the top of every runValidateCommands and read
// only by completedValidationOutcome, which the soundness invariant already
// restricts to the same process, after the same run, at the same verified tree.
var lastValidateRun validateRunRecord

// record folds one command's outcome in. Only an acceptance-CLASSIFIED command
// counts as analysis, and only a skip-driven Fail (the command itself exited 0)
// counts as a skip failure — an exit-code failure is not a skip verdict.
//
// Under the off policy nothing is parsed at all, so the record says the
// analysis was DISABLED rather than performed-and-clean. Analysed exists
// precisely to keep "ran and found nothing" distinct from "never ran".
func (r *validateRunRecord) record(cmd, out string, passed bool, status gates.Status, detail, policy string) {
	r.outputs = append(r.outputs, out)
	if !acceptance.IsExecutionCommand(cmd) {
		return
	}
	if policy == acceptance.PolicyOff {
		r.disabled = true
		return
	}
	r.analysed = true
	if passed && status == gates.Fail {
		r.failed = true
	}
	if detail != "" {
		r.notes = append(r.notes, detail)
	}
}

// judgement renders the record in the verify package's typed vocabulary.
func (r *validateRunRecord) judgement() *verify.AcceptanceJudgement {
	if r.disabled {
		return &verify.AcceptanceJudgement{
			Detail: "skip detection disabled by [validate] acceptance_skip_policy = off",
		}
	}
	if !r.analysed {
		return &verify.AcceptanceJudgement{}
	}
	detail := "no skipped, pending or undefined scenarios detected"
	if len(r.notes) > 0 {
		detail = strings.Join(r.notes, "; ")
	}
	return &verify.AcceptanceJudgement{Analysed: true, Failed: r.failed, Detail: detail}
}

// transcript is the real captured output of the commands this process ran.
func (r *validateRunRecord) transcript() string {
	return strings.TrimSpace(strings.Join(r.outputs, "\n"))
}

// describeCommands renders the record for the reused outcome's Output.
func (r *validateRunRecord) describeCommands(header string) string {
	body := r.transcript()
	if body == "" {
		return header
	}
	return fmt.Sprintf("%s\n%s", header, body)
}
