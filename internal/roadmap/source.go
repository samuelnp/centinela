package roadmap

// Source records the provenance of a deferred Backlog finding.
type Source struct {
	Feature string `json:"feature,omitempty"`
	Role    string `json:"role,omitempty"`
}
