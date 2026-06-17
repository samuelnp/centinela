package calibration

import "github.com/samuelnp/centinela/internal/config"

// Thresholds (fixed, documented constants — every recommendation cites the raw
// counts so an operator can audit and override).
const (
	highFrictionRate = 1.0  // ≥ 1.0 rework events per advance  → high friction
	lowFrictionRate  = 0.25 // ≤ 0.25 rework events per advance → low friction
	minAdvances      = 3    // need ≥ 3 advances before a rate is trustworthy
)

// strictnessRank orders profiles most→least strict: strict(2) > guided(1) >
// outcome(0). Unknown profiles rank as strict (the safe default).
func strictnessRank(profile string) int {
	switch config.NormalizeEnforcementProfile(profile) {
	case config.ProfileOutcome:
		return 0
	case config.ProfileGuided:
		return 1
	default:
		return 2
	}
}

// tighter returns the next-stricter profile, or "" if already strictest.
func tighter(profile string) string {
	switch strictnessRank(profile) {
	case 0:
		return config.ProfileGuided
	case 1:
		return config.ProfileStrict
	default:
		return ""
	}
}

// looser returns the next-looser profile, or "" if already loosest.
func looser(profile string) string {
	switch strictnessRank(profile) {
	case 2:
		return config.ProfileGuided
	case 1:
		return config.ProfileOutcome
	default:
		return ""
	}
}

// classify applies the exact classification rule for one model. It returns the
// verdict, recommendation, recommended profile, capability class, and current
// profile. A model with no capability class is Unclassified/None (never invents
// a profile). Insufficient advances or between-threshold friction → Keep.
func classify(model string, s FrictionStats, cfg *config.Config) (Verdict, Recommendation, string, string, string) {
	class, ok := config.CapabilityClassFor(model, cfg)
	if !ok {
		return Unclassified, None, "", "", ""
	}
	current := config.ProfileForCapability(class, cfg)
	if !s.HasRate || s.Advances < minAdvances {
		return WellCalibrated, Keep, current, class, current
	}
	if s.Rate >= highFrictionRate {
		if next := tighter(current); next != "" {
			return Undergoverned, Tighten, next, class, current
		}
		return WellCalibrated, Keep, current, class, current
	}
	if s.Rate <= lowFrictionRate {
		if next := looser(current); next != "" {
			return Overgoverned, Loosen, next, class, current
		}
		return WellCalibrated, Keep, current, class, current
	}
	return WellCalibrated, Keep, current, class, current
}
