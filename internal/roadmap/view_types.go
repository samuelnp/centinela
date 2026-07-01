package roadmap

// FeatureView is a derived, JSON-tagged projection of a Feature for the
// read-only `roadmap --json` contract. It is never persisted: status and
// readiness are two independent dimensions. Readiness carries a value only for
// planned features ("ready"|"blocked"); done/in-progress rows omit it and let
// status carry the signal. BlockedBy appears only on blocked rows, while
// DependsOn is always serialized (persisted input, even as []).
type FeatureView struct {
	Name      string   `json:"name"`
	Phase     string   `json:"phase"`
	Status    string   `json:"status"`
	Draft     bool     `json:"draft,omitempty"`
	Readiness string   `json:"readiness,omitempty"`
	DependsOn []string `json:"dependsOn"`
	BlockedBy []string `json:"blockedBy,omitempty"`
}

// PhaseView groups ordered FeatureViews under a phase name.
type PhaseView struct {
	Name     string        `json:"name"`
	Features []FeatureView `json:"features"`
}

// StatusCounts tallies schedulable features by status; Backlog and Baseline
// entries are excluded, matching Summary() scoping.
type StatusCounts struct {
	Planned    int `json:"planned"`
	InProgress int `json:"inProgress"`
	Done       int `json:"done"`
}

// RoadmapView is the deterministic, machine-readable projection of a Roadmap
// emitted by `roadmap --json`. Ordered slices only, for byte-stable output.
type RoadmapView struct {
	Phases []PhaseView  `json:"phases"`
	Counts StatusCounts `json:"counts"`
}
