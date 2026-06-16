package audit

import "github.com/samuelnp/centinela/internal/config"

// defaultParticipants are the gates that emit per-violation Details and can
// therefore be baselined per-violation (Decision #5). Gates whose only Detail is
// a summary message (e.g. roadmap_drift) are excluded.
var defaultParticipants = []string{
	"G1: File Size",
	"import_graph",
	"spec-traceability-gate",
	"G-Secrets: Secret Scan",
	"G11: i18n",
}

// participatingGates returns the set of gate Names whose violations the ratchet
// tracks: the default detail-emitting set, intersected with target_gates when
// that allowlist is non-empty. An empty allowlist means "all defaults".
func participatingGates(cfg *config.Config) map[string]bool {
	allow := allowSet(cfg.Gates.AuditBaseline.TargetGates)
	out := make(map[string]bool, len(defaultParticipants))
	for _, name := range defaultParticipants {
		if allow == nil || allow[name] {
			out[name] = true
		}
	}
	return out
}

// isParticipating reports whether a single gate Name participates in the ratchet.
func isParticipating(name string, cfg *config.Config) bool {
	return participatingGates(cfg)[name]
}

// allowSet builds a lookup from the configured target_gates; nil means "no
// restriction" (every default participates).
func allowSet(targets []string) map[string]bool {
	if len(targets) == 0 {
		return nil
	}
	set := make(map[string]bool, len(targets))
	for _, t := range targets {
		set[t] = true
	}
	return set
}
