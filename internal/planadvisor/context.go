package planadvisor

import "github.com/samuelnp/centinela/internal/config"

type bundle struct {
	Feature                string
	Coverage               coverage
	Dependencies, Siblings []string
	Lessons, QualityNotes  []string
	Memory                 []string
}

func buildBundle(feature string, cfg *config.Config) bundle {
	a := loadArtifacts(feature)
	deps, sibs := relatedNames(feature)
	related := append(append([]string{}, deps...), sibs...)
	return bundle{
		Feature:      feature,
		Coverage:     scanTexts(a),
		Dependencies: deps,
		Siblings:     sibs,
		Lessons:      relatedLessons(related),
		QualityNotes: relatedQualityNotes(feature, related),
		Memory:       recalledMemory(feature, deps, cfg),
	}
}
