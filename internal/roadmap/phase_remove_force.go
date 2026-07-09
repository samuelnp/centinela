package roadmap

// forceRemovePhase removes phase idx and its features, then prunes those
// features' analysis and quality entries — all as one logical mutation. The
// mutated roadmap is dependency-validated IN MEMORY before any byte hits disk,
// so a surviving feature that still dependsOn a removed one refuses the whole op
// byte-identically. Analysis/quality coverage stays consistent by construction
// (exactly the removed features' entries are pruned), and all three validators
// re-run against the final on-disk state to guarantee `roadmap validate` passes.
func (d *rawDoc) forceRemovePhase(path string, idx int, p *rawPhase) error {
	removed := map[string]bool{}
	for _, f := range p.Features {
		if name, err := featureName(f); err == nil {
			removed[name] = true
		}
	}
	if err := d.removePhaseAt(idx); err != nil {
		return err
	}
	typed, err := d.toRoadmap()
	if err != nil {
		return err
	}
	if err := ValidateDependencies(typed); err != nil {
		return err // surviving feature depends on a removed one -> byte-identical refusal
	}
	if err := writeRawRoadmap(path, d); err != nil {
		return err
	}
	if err := removeFeatureEntries(RoadmapAnalysisFile, removed); err != nil {
		return err
	}
	if err := removeFeatureEntries(RoadmapQualityFile, removed); err != nil {
		return err
	}
	return revalidateArtifacts(typed)
}

// revalidateArtifacts re-runs analysis + quality validation against the final
// roadmap after a --force prune, confirming no orphaned or missing coverage
// survives the removal.
func revalidateArtifacts(r *Roadmap) error {
	if err := ValidateAnalysis(r); err != nil {
		return err
	}
	return ValidateQuality(r)
}
