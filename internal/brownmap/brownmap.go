// Package brownmap is an aggregator that turns the read-only analyze Inventory
// into a DRAFT roadmap for a brownfield repo: it regenerates the in-process
// reconstruct.Reconstruction and partitions capability into a Baseline phase
// (already-built surfaces, never re-planned) plus net-new gap phase(s) seeded
// from reconstruct "# TODO: confirm" targets and repeatable --goal strings. It
// is deterministic (no LLM), byte-stable over reconstruct's already-sorted
// Targets, and NEVER clobbers the canonical .workflow/roadmap.json — WriteDraft
// emits an atomic draft file. It imports the internal/analyze (domain),
// internal/roadmap (domain) and internal/reconstruct (aggregator) packages
// read-only plus stdlib; it is invoked from cmd/ and its Plan type is rendered
// by internal/ui.
package brownmap

import (
	"github.com/samuelnp/centinela/internal/analyze"
	"github.com/samuelnp/centinela/internal/roadmap"
)

// DefaultDraftPath is the well-known draft location the brownfield roadmap is
// written to. It is deliberately NOT roadmap.RoadmapFile so a curated roadmap is
// never clobbered; it lives here (not in cmd/) so callers share the constant.
const DefaultDraftPath = ".workflow/roadmap.brownfield.json"

// Plan is the typed, byte-stable result of a brownfield draft generation: the
// assembled draft Roadmap (Baseline phase + gap phase(s)), the count of Baseline
// entries and gap features, and the draft path it was/will be written to.
type Plan struct {
	Roadmap       roadmap.Roadmap
	BaselineCount int
	GapCount      int
	DraftPath     string
}

// Brownfielder maps an Inventory plus user-stated goals to a draft Plan. It is
// an interface so a future LLM-backed generator can replace the deterministic
// default without touching cmd/ or the writer, mirroring reconstruct.Reconstructor.
type Brownfielder interface {
	Generate(inv analyze.Inventory, goals []string) Plan
}

type ruleBrownfielder struct{}

// NewBrownfielder returns the default deterministic, rule-driven brownfielder.
func NewBrownfielder() Brownfielder { return ruleBrownfielder{} }
