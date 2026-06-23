// Package reconstruct is an aggregator that consumes the read-only
// internal/analyze Inventory contract to deterministically reconstruct a
// behavioral spec corpus: one specs/<slug>.feature Gherkin skeleton plus one
// docs/features/<slug>.md brief stub per significant module/surface, with honest
// "# TODO: confirm" gaps. It is deterministic (no LLM), byte-stable, and never
// clobbers hand-authored specs. It imports only the internal/analyze domain
// package and the standard library; it is invoked from cmd/ and its types are
// rendered by internal/ui.
package reconstruct

import "strings"

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

// TodoTargets returns, in the Reconstruction's already-sorted target order, the
// targets whose assembled feature artifact still carries an un-confirmed
// "# TODO: confirm" marker — i.e. the surfaces whose reconstructed behavior is
// not yet confirmed. It is a thin read-only accessor over the existing
// Features/Targets so consumers (e.g. internal/brownmap) need not duplicate the
// skeleton's TODO-marker rule. A target with no TODO-bearing artifact is omitted.
func (r Reconstruction) TodoTargets() []Target {
	bySlug := map[string]bool{}
	for _, f := range r.Features {
		if strings.Contains(f.Body, todoMarker) {
			bySlug[f.Slug] = true
		}
	}
	var out []Target
	for _, t := range r.Targets {
		if bySlug[t.Slug] {
			out = append(out, t)
		}
	}
	return out
}
