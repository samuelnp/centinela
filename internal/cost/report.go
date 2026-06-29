package cost

import (
	"sort"

	"github.com/samuelnp/centinela/internal/config"
)

// Report is the rendered cost picture: spend vs budget per feature, per
// feature/step, and per model. Deterministically ordered by Name.
type Report struct {
	Features []Status `json:"features"`
	Steps    []Status `json:"steps"`
	Models   []Status `json:"models"`
}

// Build folds the aggregate against the configured budgets into a Report.
func Build(a Aggregate, cfg config.CostConfig) Report {
	var r Report
	for f, u := range a.Feature {
		r.Features = append(r.Features, status("feature", f, u.Tokens(), cfg.FeatureTokenBudget))
	}
	for f, steps := range a.Step {
		for s, u := range steps {
			r.Steps = append(r.Steps, status("step", f+"/"+s, u.Tokens(), cfg.StepTokenBudget))
		}
	}
	for m, u := range a.Model {
		r.Models = append(r.Models, status("model", m, u.Tokens(), cfg.TierBudgets[m]))
	}
	sortByName(r.Features)
	sortByName(r.Steps)
	sortByName(r.Models)
	return r
}

// AnyOver reports whether any scope exceeded its budget.
func (r Report) AnyOver() bool {
	for _, set := range [][]Status{r.Features, r.Steps, r.Models} {
		for _, s := range set {
			if s.Over {
				return true
			}
		}
	}
	return false
}

// Empty reports whether no spend has been recorded at all.
func (r Report) Empty() bool {
	return len(r.Features) == 0 && len(r.Steps) == 0 && len(r.Models) == 0
}

// ActiveStatus returns the over-budget Status for the active feature/step — the
// step budget first, then the feature budget — for the soft-gate warning. The
// second return is false when neither scope is over budget.
func ActiveStatus(a Aggregate, cfg config.CostConfig, feature, step string) (Status, bool) {
	if su, ok := a.Step[feature][step]; ok {
		if st := status("step", feature+"/"+step, su.Tokens(), cfg.StepTokenBudget); st.Over {
			return st, true
		}
	}
	if fu, ok := a.Feature[feature]; ok {
		if st := status("feature", feature, fu.Tokens(), cfg.FeatureTokenBudget); st.Over {
			return st, true
		}
	}
	return Status{}, false
}

func sortByName(s []Status) {
	sort.Slice(s, func(i, j int) bool { return s[i].Name < s[j].Name })
}
