package roadmap

// Feature is a single deliverable within a phase.
type Feature struct {
	Name      string   `json:"name"`
	DependsOn []string `json:"dependsOn,omitempty"`
	// Archetype optionally pins the workflow track for this feature; an explicit
	// --archetype flag at start overrides it. Empty resolves to canonical.
	Archetype string `json:"archetype,omitempty"`
	// Description is the human-facing bullet prose rendered after the em-dash in
	// ROADMAP.md. Optional; preserved through rawio round-trips as typed JSON.
	Description string `json:"description,omitempty"`
	// Fixes is the "*Fixes: …*" clause rendered on its own indented line.
	Fixes      string  `json:"fixes,omitempty"`
	Summary    string  `json:"summary,omitempty"`    // deferred-finding one-liner
	Source     *Source `json:"source,omitempty"`     // {feature, role} provenance
	DeferredAt string  `json:"deferredAt,omitempty"` // RFC3339 capture time
	// Draft marks an authored-but-unscored feature living in a schedulable phase.
	// A draft is exempt from the ≥9 analysis/quality coverage set (via the single
	// NonBacklogFeatureSet hook), classifies as State:"draft" (not ready), is not
	// counted as committed work in Summary, serializes readiness:"draft" in the
	// JSON view, and refuses `start` until finalized by `roadmap promote`.
	Draft bool `json:"draft,omitempty"`
}

// Phase groups related features under a milestone.
type Phase struct {
	Name string `json:"name"`
	// Note is the optional blockquote rationale rendered after the phase heading
	// and before the feature list. May be multi-paragraph (\n\n separated).
	Note     string    `json:"note,omitempty"`
	Features []Feature `json:"features"`
}

// Roadmap holds the full project plan.
type Roadmap struct {
	// Intro is the optional top-of-file blockquote rendered after "# Roadmap".
	Intro  string  `json:"intro,omitempty"`
	Phases []Phase `json:"phases"`
}
