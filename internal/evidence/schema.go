package evidence

import (
	"bytes"
	"encoding/json"
)

// Meta carries CLI-tagged provenance. It is OPTIONAL on read (older files
// predating this CLI lack _meta) but always emitted on write.
type Meta struct {
	CLIVersion string `json:"cli_version"`
	WrittenAt  string `json:"written_at"`
}

// RoleEvidence is the in-memory representation of one
// .workflow/<feature>-<role>.json artifact. It mirrors the contract in
// docs/architecture/evidence-contract.md verbatim. Extra preserves any
// fields the current binary does not know about so a newer file can be
// round-tripped through an older binary without data loss (AC7).
type RoleEvidence struct {
	Meta        *Meta    `json:"_meta,omitempty"`
	Feature     string   `json:"feature"`
	Step        string   `json:"step"`
	Role        string   `json:"role"`
	Status      string   `json:"status"`
	GeneratedAt string   `json:"generatedAt"`
	Inputs      []string `json:"inputs"`
	Outputs     []string `json:"outputs"`
	EdgeCases   []string `json:"edgeCases"`
	MobileFirst *bool    `json:"mobileFirst,omitempty"`
	// Coverage is the claimed per-package coverage percentage (e.g. 85.0 for
	// 85%). Optional: a nil pointer means the role made no coverage claim and
	// the coverage check is skipped. Typed to forbid free-form prose claims.
	Coverage  *float64                   `json:"coverage,omitempty"`
	HandoffTo string                     `json:"handoffTo"`
	Extra     map[string]json.RawMessage `json:"-"`
}

// jsonKeyOrder is the canonical key order on disk. Used by MarshalJSON.
var jsonKeyOrder = []string{
	"_meta", "feature", "step", "role", "status", "generatedAt",
	"inputs", "outputs", "edgeCases", "mobileFirst", "coverage", "handoffTo",
}

// MarshalJSON emits a stable key order so two writes of the same evidence
// produce byte-identical output (the postwrite hook in Slice 2 relies on
// this). Unknown fields land last, sorted alphabetically.
func (r *RoleEvidence) MarshalJSON() ([]byte, error) {
	return encodeOrdered(r.knownFieldsMap(), r.Extra)
}

func encodeOrdered(known map[string]json.RawMessage, extra map[string]json.RawMessage) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	first := true
	emit := func(k string, v json.RawMessage) {
		if !first {
			buf.WriteByte(',')
		}
		first = false
		keyBytes, _ := json.Marshal(k)
		buf.Write(keyBytes)
		buf.WriteByte(':')
		buf.Write(v)
	}
	for _, k := range jsonKeyOrder {
		if v, ok := known[k]; ok {
			emit(k, v)
		}
	}
	for _, k := range sortedKeys(extra) {
		emit(k, extra[k])
	}
	buf.WriteByte('}')
	return prettyIndent(buf.Bytes())
}
