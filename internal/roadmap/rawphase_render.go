package roadmap

import (
	"bytes"
	"encoding/json"
	"sort"
)

// renderDirtyPhase renders one phase at the "    " phase indent, emitting its
// "features" array one compact object per line so concurrent appends conflict
// as a trivial textual union and a single-feature edit is a single-line diff.
// Every phase key is preserved: name and note keep their authored position,
// any other key follows in sorted order, and "features" is emitted last.
func renderDirtyPhase(raw json.RawMessage) (string, error) {
	var phase map[string]json.RawMessage
	if err := json.Unmarshal(raw, &phase); err != nil {
		return "", err
	}
	var feats []json.RawMessage
	if f, ok := phase["features"]; ok {
		if err := json.Unmarshal(f, &feats); err != nil {
			return "", err
		}
	}
	var buf bytes.Buffer
	buf.WriteString("{")
	first := true
	for _, key := range phaseKeyOrder(phase) {
		writePhaseKey(&buf, key, phase[key], &first)
	}
	// "features" emitted last, one object per line.
	if !first {
		buf.WriteByte(',')
	}
	buf.WriteString("\n      \"features\": [")
	for i, f := range feats {
		buf.WriteString("\n        ")
		var c bytes.Buffer
		_ = json.Compact(&c, f)
		buf.Write(c.Bytes())
		if i < len(feats)-1 {
			buf.WriteByte(',')
		}
	}
	if len(feats) > 0 {
		buf.WriteString("\n      ")
	}
	buf.WriteString("]\n    }")
	return buf.String(), nil
}

func writePhaseKey(buf *bytes.Buffer, key string, v json.RawMessage, first *bool) {
	if !*first {
		buf.WriteByte(',')
	}
	*first = false
	k, _ := json.Marshal(key)
	buf.WriteString("\n      ")
	buf.Write(k)
	buf.WriteString(": ")
	var c bytes.Buffer
	_ = json.Compact(&c, v)
	buf.Write(c.Bytes())
}

// phaseKeyOrder lists every phase key except "features" (emitted last): the
// authored leaders first, then the rest sorted so an unknown key round-trips
// in a stable position rather than at Go's randomized map order.
func phaseKeyOrder(phase map[string]json.RawMessage) []string {
	order := make([]string, 0, len(phase))
	for _, key := range []string{"name", "note"} {
		if _, ok := phase[key]; ok {
			order = append(order, key)
		}
	}
	var extra []string
	for key := range phase {
		if key == "name" || key == "note" || key == "features" {
			continue
		}
		extra = append(extra, key)
	}
	sort.Strings(extra)
	return append(order, extra...)
}
