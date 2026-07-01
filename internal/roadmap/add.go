package roadmap

// AddRequest carries a resolved `roadmap add` request. Phase must name a
// schedulable (non-Backlog, non-Baseline) phase; the feature is recorded as a
// draft so `roadmap validate` stays PASS until it is scored via promote.
type AddRequest struct {
	Slug        string
	Phase       string
	Description string
	Archetype   string
	DependsOn   []string
}

// Add appends a new draft feature to a schedulable phase via raw-preserving
// read-modify-write. Every validation (slug shape, no collision, known phase,
// dependency integrity) runs before the single atomic write, so a rejected add
// leaves roadmap.json byte-identical.
func Add(path string, req AddRequest) error {
	if err := validateSlug(req.Slug); err != nil {
		return err
	}
	doc, err := readRawRoadmap(path)
	if err != nil {
		return err
	}
	existing, err := doc.phaseFeatureNames()
	if err != nil {
		return err
	}
	if err := validateNoCollision(req.Slug, existing); err != nil {
		return err
	}
	deps := req.DependsOn
	if deps == nil {
		deps = []string{}
	}
	entry, err := compactBytes(Feature{
		Name: req.Slug, DependsOn: deps, Archetype: req.Archetype,
		Description: req.Description, Draft: true,
	})
	if err != nil {
		return err
	}
	if err := doc.appendFeatureToPhase(req.Phase, entry); err != nil {
		return err
	}
	typed, err := doc.toRoadmap()
	if err != nil {
		return err
	}
	if err := ValidateDependencies(typed); err != nil {
		return err
	}
	return writeRawRoadmap(path, doc)
}
