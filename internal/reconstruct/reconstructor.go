package reconstruct

import "github.com/samuelnp/centinela/internal/analyze"

// Reconstructor maps an Inventory to a Reconstruction. It is an interface so a
// future LLM-backed reconstructor can replace the deterministic default without
// touching cmd/ or the writer, mirroring synthesize.Inferer.
type Reconstructor interface {
	Reconstruct(inv analyze.Inventory) Reconstruction
}

type ruleReconstructor struct{}

// NewReconstructor returns the default deterministic, rule-table-driven
// reconstructor.
func NewReconstructor() Reconstructor { return ruleReconstructor{} }

// Reconstruct runs Select, then for each sorted target assembles the pure-string
// feature skeleton and brief stub and tallies the total TODO markers. It
// performs no I/O and is byte-stable for a given Inventory (sorted targets,
// stable section order, no map-iteration order in the output).
func (ruleReconstructor) Reconstruct(inv analyze.Inventory) Reconstruction {
	targets := Select(inv)
	r := Reconstruction{Targets: targets}
	for _, t := range targets {
		body, todos := featureSkeleton(t)
		r.Features = append(r.Features, Artifact{Slug: t.Slug, Body: body})
		r.Briefs = append(r.Briefs, Artifact{Slug: t.Slug, Body: briefStub(t)})
		r.TodoCount += todos
	}
	return r
}
