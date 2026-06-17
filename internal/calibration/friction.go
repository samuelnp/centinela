package calibration

import "github.com/samuelnp/centinela/internal/telemetry"

// unattributed is the bucket key for events with no Model (back-compat old lines
// and any emit with no resolvable driver model).
const unattributed = "unattributed"

// modelKey maps an event's Model to its bucket key, folding empty into the
// single "unattributed" bucket.
func modelKey(model string) string {
	if model == "" {
		return unattributed
	}
	return model
}

// reworkType reports whether an event type counts toward rework (governance
// friction before green): gate-failure + verify-rejection + complete-rejected.
// step-advanced is excluded. Mirrors insights.reworkType.
func reworkType(t string) bool {
	switch t {
	case telemetry.TypeGateFailure,
		telemetry.TypeVerifyRejection,
		telemetry.TypeCompleteRejected:
		return true
	default:
		return false
	}
}

// frictionByModel tallies per-model FrictionStats from the event slice, guarding
// the rate denominator (HasRate=false when Advances==0; never divides by zero).
func frictionByModel(events []telemetry.Event) map[string]FrictionStats {
	out := make(map[string]FrictionStats)
	for _, e := range events {
		k := modelKey(e.Model)
		s := out[k]
		switch e.Type {
		case telemetry.TypeBlock:
			s.Blocks++
		case telemetry.TypeGateFailure:
			s.GateFailures++
		case telemetry.TypeVerifyRejection:
			s.VerifyRejections++
		case telemetry.TypeStepAdvanced:
			s.Advances++
		}
		if reworkType(e.Type) {
			s.Rework++
		}
		out[k] = s
	}
	for k, s := range out {
		if s.Advances > 0 {
			s.Rate = float64(s.Rework) / float64(s.Advances)
			s.HasRate = true
			out[k] = s
		}
	}
	return out
}
