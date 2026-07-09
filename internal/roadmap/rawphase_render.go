package roadmap

import (
	"bytes"
	"encoding/json"
)

// renderDirtyPhase renders a mutated phase (Backlog or promote target) at the
// "    " phase indent, emitting its "features" array one compact object per
// line so concurrent appends conflict as a trivial textual union. The phase's
// name (and any other top-level phase keys) are preserved.
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
	for _, key := range []string{"name", "note"} {
		if v, ok := phase[key]; ok {
			writePhaseKey(&buf, key, v, &first)
		}
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
