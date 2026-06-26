package teamdashboard

// gateTopN bounds the gate-health ranking. It is pinned as a package const so
// the board's failure counts never diverge from `centinela insights`, which
// ranks the same gate-failure events via insights.Gates.
const gateTopN = 10

// Compute turns Inputs into a pure Dashboard. Empty/nil sources each yield an
// honest empty state, never a panic: no Active -> empty Features; nil Roadmap
// -> RoadmapBurndown{Present:false}; no gate-failure events -> empty Gates.
func Compute(in Inputs) Dashboard {
	return Dashboard{
		Features: features(in.Active, in.Owners, in.Now),
		Roadmap:  burndown(in.Roadmap),
		Gates:    gatehealth(in.Events, gateTopN),
	}
}
