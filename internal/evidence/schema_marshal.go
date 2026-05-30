package evidence

import (
	"bytes"
	"encoding/json"
	"sort"
)

// marshalNoEscape mirrors json.Marshal but disables HTML escaping so the
// evidence files stay human-readable (placeholders like <feature-slug> do
// not get rewritten to <>).
func marshalNoEscape(v any) (json.RawMessage, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	out := buf.Bytes()
	// json.Encoder appends a trailing newline.
	if n := len(out); n > 0 && out[n-1] == '\n' {
		out = out[:n-1]
	}
	return json.RawMessage(out), nil
}

// knownFieldsMap renders every contract-defined field, honouring omitempty
// for _meta and mobileFirst exactly as MarshalJSON would on individual
// fields. The map is then re-ordered by jsonKeyOrder in MarshalJSON.
// All values are primitives or small structs whose JSON encoding cannot
// fail at runtime, so the marshalNoEscape calls are safe to err-swallow.
func (r *RoleEvidence) knownFieldsMap() map[string]json.RawMessage {
	out := map[string]json.RawMessage{}
	put := func(key string, v any) {
		raw, _ := marshalNoEscape(v)
		out[key] = raw
	}
	if r.Meta != nil {
		put("_meta", r.Meta)
	}
	if r.MobileFirst != nil {
		put("mobileFirst", r.MobileFirst)
	}
	if r.Coverage != nil {
		put("coverage", r.Coverage)
	}
	put("feature", r.Feature)
	put("step", r.Step)
	put("role", r.Role)
	put("status", r.Status)
	put("generatedAt", r.GeneratedAt)
	put("inputs", nonNilStrings(r.Inputs))
	put("outputs", nonNilStrings(r.Outputs))
	put("edgeCases", nonNilStrings(r.EdgeCases))
	put("handoffTo", r.HandoffTo)
	return out
}

func nonNilStrings(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func sortedKeys(m map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// prettyIndent re-formats raw with two-space indentation so files on disk
// are diff-friendly.
func prettyIndent(raw []byte) ([]byte, error) {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, raw, "", "  "); err != nil {
		return nil, err
	}
	pretty.WriteByte('\n')
	return pretty.Bytes(), nil
}
