package main

import "github.com/samuelnp/centinela/internal/evidence"

// invalidateDownstream sheds the certification evidence of every re-opened step
// so the next complete re-runs its gates. The per-step role/artifact policy
// lives in internal/evidence (G7); this composes it — looping, deduping across
// steps, and counting what was actually removed. Idempotent: already-absent
// artifacts are not errors and do not count.
func invalidateDownstream(feature string, steps []string) (int, error) {
	count := 0
	seenRole := map[evidence.Role]bool{}
	seenArtifact := map[string]bool{}
	for _, step := range steps {
		roles, artifacts := evidence.InvalidationTargets(feature, step)
		for _, role := range roles {
			if seenRole[role] {
				continue
			}
			seenRole[role] = true
			removed, err := evidence.Invalidate(feature, role)
			if err != nil {
				return count, err
			}
			if removed {
				count++
			}
		}
		for _, suffix := range artifacts {
			if seenArtifact[suffix] {
				continue
			}
			seenArtifact[suffix] = true
			removed, err := evidence.InvalidateArtifact(feature, suffix)
			if err != nil {
				return count, err
			}
			if removed {
				count++
			}
		}
	}
	return count, nil
}
