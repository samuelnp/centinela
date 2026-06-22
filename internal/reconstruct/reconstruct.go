// Package reconstruct is an aggregator that consumes the read-only
// internal/analyze Inventory contract to deterministically reconstruct a
// behavioral spec corpus: one specs/<slug>.feature Gherkin skeleton plus one
// docs/features/<slug>.md brief stub per significant module/surface, with honest
// "# TODO: confirm" gaps. It is deterministic (no LLM), byte-stable, and never
// clobbers hand-authored specs. It imports only the internal/analyze domain
// package and the standard library; it is invoked from cmd/ and its types are
// rendered by internal/ui.
package reconstruct

// Role is the inferred behavioral kind of a target, driving role-aware scenario
// skeletons. An unknown role still yields a non-empty, assertion-free stub.
type Role string

const (
	// RoleCommand is a CLI/command surface (invocation scenario).
	RoleCommand Role = "command"
	// RoleEndpoint is an HTTP/RPC surface (request/response scenario).
	RoleEndpoint Role = "endpoint"
	// RoleModule is a behavioral package with no more specific role.
	RoleModule Role = "module"
)

// Target is one selected surface the reconstructor will emit artifacts for. Pkg
// is the source package path; Slug is its deterministic, collision-disambiguated
// filename stem; Role drives the skeleton shape; Reason records why it was
// promoted, for auditability.
type Target struct {
	Pkg    string `json:"pkg"`
	Slug   string `json:"slug"`
	Role   Role   `json:"role"`
	Reason string `json:"reason"`
}

// Reconstruction is the typed, byte-stable result of a reconstruct run: the
// sorted targets, the assembled feature/brief corpus keyed by slug, the total
// "# TODO: confirm" markers across all features, and the paths written/skipped
// once WriteCorpus runs.
type Reconstruction struct {
	Targets   []Target   `json:"targets"`
	Features  []Artifact `json:"-"`
	Briefs    []Artifact `json:"-"`
	TodoCount int        `json:"todoCount"`
	Written   []string   `json:"written"`
	Skipped   []string   `json:"skipped"`
}

// Artifact is one assembled file: its slug-derived basename and pure-string body.
type Artifact struct {
	Slug string
	Body string
}
