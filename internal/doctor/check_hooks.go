package doctor

import (
	"os"

	"github.com/samuelnp/centinela/internal/setup"
)

const claudeSettingsPath = ".claude/settings.json"

// hooksCheck verifies that centinela's prewrite/postwrite/context/statusline
// (and OpenCode-equivalent) hook entries are wired into .claude/settings.json.
// It reuses setup.BuildSyncPlan, whose plan is non-empty precisely when a
// re-wire would change something (i.e. an entry is missing or stale).
type hooksCheck struct{}

func (hooksCheck) Name() string { return "hooks" }

func (hooksCheck) Run(Context) Diagnosis {
	d := Diagnosis{Name: "hooks"}
	if _, err := os.Stat(".claude"); os.IsNotExist(err) {
		d.Status = Warn
		d.Message = "no .claude/ directory — hooks not configured; run `centinela setup`"
		return d
	}
	plan, err := setup.BuildSyncPlan("both")
	if err != nil {
		d.Status = Error
		d.Message = "cannot inspect hook settings: " + err.Error()
		return d
	}
	if !plan.HasChanges() {
		d.Status = OK
		d.Message = "all centinela hook entries are wired"
		return d
	}
	d.Status = Error
	d.Message = "centinela hook entries are missing or stale in " + claudeSettingsPath
	for _, it := range plan.Items {
		d.Details = append(d.Details, string(it.Kind)+" needs "+string(it.Action)+" at "+it.Path)
	}
	d.Repair = &Repair{
		Safe:       true,
		Idempotent: true,
		Apply: func() error {
			p, err := setup.BuildSyncPlan("both")
			if err != nil {
				return err
			}
			return setup.ApplySync(p)
		},
	}
	return d
}
