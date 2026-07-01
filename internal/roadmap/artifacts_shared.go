package roadmap

// appendScoreArtifacts records a scored feature in the analysis + quality JSON
// (raw-preserving, one object per line) and appends the given provenance bullet
// to their .md companions. Shared by both promote branches — the Backlog move
// and the in-place draft finalize — so they stay DRY and under the line budget.
// Each underlying write is atomic (temp-file+rename).
func appendScoreArtifacts(slug, summary string, scores QualityScores, bullet string) error {
	analysisEntry, err := compactBytes(AnalysisFeature{Name: slug})
	if err != nil {
		return err
	}
	if err := appendFeatureEntry(RoadmapAnalysisFile, analysisEntry); err != nil {
		return err
	}
	qualityEntry, err := compactBytes(QualityFeature{Name: slug, Scores: scores, Summary: summary})
	if err != nil {
		return err
	}
	if err := appendFeatureEntry(RoadmapQualityFile, qualityEntry); err != nil {
		return err
	}
	if err := appendLine(RoadmapAnalysisMarkdown, bullet); err != nil {
		return err
	}
	return appendLine(RoadmapQualityMarkdown, bullet)
}
