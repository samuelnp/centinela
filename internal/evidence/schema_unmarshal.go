package evidence

import (
	"encoding/json"
	"fmt"
)

// knownKeys is the set of fields parsed into typed RoleEvidence fields.
// Anything else lands in Extra to survive round-trips through older binaries.
var knownKeys = func() map[string]struct{} {
	m := map[string]struct{}{}
	for _, k := range jsonKeyOrder {
		m[k] = struct{}{}
	}
	return m
}()

// UnmarshalJSON parses raw into r while preserving unknown fields in Extra.
func (r *RoleEvidence) UnmarshalJSON(data []byte) error {
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("evidence: invalid json: %w", err)
	}
	if err := r.assignKnown(raw); err != nil {
		return err
	}
	r.Extra = map[string]json.RawMessage{}
	for k, v := range raw {
		if _, ok := knownKeys[k]; ok {
			continue
		}
		r.Extra[k] = v
	}
	return nil
}

func (r *RoleEvidence) assignKnown(raw map[string]json.RawMessage) error {
	type binding struct {
		key string
		dst any
	}
	bindings := []binding{
		{"_meta", &r.Meta}, {"feature", &r.Feature}, {"step", &r.Step},
		{"role", &r.Role}, {"status", &r.Status}, {"generatedAt", &r.GeneratedAt},
		{"inputs", &r.Inputs}, {"outputs", &r.Outputs}, {"edgeCases", &r.EdgeCases},
		{"mobileFirst", &r.MobileFirst}, {"handoffTo", &r.HandoffTo},
	}
	for _, b := range bindings {
		v, ok := raw[b.key]
		if !ok {
			continue
		}
		if err := json.Unmarshal(v, b.dst); err != nil {
			return fmt.Errorf("evidence: invalid %s: %w", b.key, err)
		}
	}
	return nil
}
