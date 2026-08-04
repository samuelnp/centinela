package doctor

import (
	"path/filepath"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// profileDefaultCheck tells a project that it is inheriting the shipped
// enforcement-profile default rather than a choice it made. The default is now
// guided, and a project that predates the flip may still expect strict.
//
// ADVISORY BY CONSTRUCTION: it reports Warn — the doctor severity documented as
// "does not fail the command" — and never Error, so ExitError stays false on its
// account. It carries no Repair, so `--fix` cannot rewrite a user's
// centinela.toml behind their back; the exact line to add is in the details.
//
// It stays silent on a project with no workflows (nothing has inherited
// anything yet) and on a project whose centinela.toml pins the profile
// explicitly, whatever value it pins.
//
// A config that could not be PARSED is not the same as one with no pin. NewContext
// leaves ctx.Config nil on a parse error so the other checks keep running, and
// reading that nil as "unpinned" told operators whose config pins strict that they
// were inheriting guided — the doctor, of all surfaces, answering with the loosest
// profile. When the config is unreadable this check says so and defers to the
// config check, which already reports the parse failure as an ERROR.
type profileDefaultCheck struct{}

func (profileDefaultCheck) Name() string { return "profile-default" }

func (profileDefaultCheck) Run(ctx Context) Diagnosis {
	d := Diagnosis{Name: "profile-default"}
	if ctx.CfgErr != nil || ctx.Config == nil {
		d.Status = OK
		d.Message = "enforcement profile is unknown — " + config.Filename + " could not be read"
		return d
	}
	// Silent unless the project ACTUALLY inherits guided. An explicit pin is not
	// the only way to land elsewhere: a declared driver model resolves through the
	// capability tier, and an unmapped one resolves strict. Warning "you are on
	// guided" there would repeat, at project scope, the exact mistake this check
	// was hardened against at parse-error scope.
	inherited := config.ProjectDefaultProfile(ctx.Config)
	pinned := ctx.Config.Workflow.RawEnforcementProfile != ""
	if pinned || inherited != config.ProfileGuided || !hasWorkflows(ctx.Root) {
		d.Status = OK
		d.Message = "enforcement profile is pinned, resolves elsewhere, or no workflows exist"
		return d
	}
	d.Status = Warn
	d.Message = "enforcement profile is inherited, and the default is now " + config.ProfileGuided
	d.Details = []string{
		"guided relaxes PROCESS only — confirmation cadence, the orchestration " +
			"evidence bundle and the greenfield setup cascade",
		"every gate, the full test suite, verification freshness, claim " +
			"verification and the verifier's grounded verdict are unchanged",
		`to keep the previous behavior, add to centinela.toml: [workflow] enforcement_profile = "strict"`,
		"workflows started before the flip are unaffected — they stay strict for life",
		"this is the PROJECT default; a workflow that pinned a driver model or a " +
			"--profile at start resolves on its own terms (see centinela status)",
	}
	return d
}

// hasWorkflows reports whether any feature workflow state exists under
// .workflow/, which is what makes an inherited default consequential.
func hasWorkflows(root string) bool {
	return len(workflow.ActiveWorkflows(filepath.Join(root, ".workflow"))) > 0
}
