package hookpolicy

import (
	"bytes"
	"encoding/json"
	"sort"
)

// jsonKeyOrder mirrors internal/evidence.jsonKeyOrder. Duplicated rather
// than imported to keep the hookpolicy layer dependency-thin (it must not
// pull internal/evidence which already imports it for path helpers in
// neighbouring builds). Drift is caught by format_evidence_test.go which
// compares against the canonical order produced by the evidence package.
var jsonKeyOrder = []string{
	"_meta", "feature", "step", "role", "status", "generatedAt",
	"inputs", "outputs", "edgeCases", "mobileFirst", "handoffTo",
}

// encodeOrderedEvidence serialises a parsed map back to JSON with the
// canonical evidence key order. Unknown keys land last, sorted
// alphabetically. The output is compact; the caller indents. Cannot fail
// at runtime: every value already arrived as `json.RawMessage`, and key
// marshalling of a Go string never errors.
func encodeOrderedEvidence(m map[string]json.RawMessage) []byte {
	known := map[string]struct{}{}
	for _, k := range jsonKeyOrder {
		known[k] = struct{}{}
	}
	extras := make([]string, 0, len(m))
	for k := range m {
		if _, ok := known[k]; !ok {
			extras = append(extras, k)
		}
	}
	sort.Strings(extras)
	var buf bytes.Buffer
	buf.WriteByte('{')
	first := true
	emit := func(k string, v json.RawMessage) {
		if !first {
			buf.WriteByte(',')
		}
		first = false
		kb, _ := json.Marshal(k)
		buf.Write(kb)
		buf.WriteByte(':')
		buf.Write(v)
	}
	for _, k := range jsonKeyOrder {
		if v, ok := m[k]; ok {
			emit(k, v)
		}
	}
	for _, k := range extras {
		emit(k, m[k])
	}
	buf.WriteByte('}')
	return buf.Bytes()
}
