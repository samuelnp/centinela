package roadmap

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// render serializes the doc back to bytes at 2-space indent. EVERY phase —
// not only the mutated one — renders through renderDirtyPhase's canonical
// one-feature-object-per-line form, so a one-field edit produces a one-line
// diff instead of reformatting a whole phase (AC14). Rendering is idempotent:
// a second write of unchanged content is byte-identical, and unknown
// per-phase and per-feature fields survive the round trip verbatim.
func (d *rawDoc) render() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("{\n  \"phases\": [\n")
	for i := range d.phases {
		out, err := renderDirtyPhase(d.phaseBytes(i))
		if err != nil {
			return nil, err
		}
		buf.WriteString("    ")
		buf.WriteString(out)
		if i < len(d.phases)-1 {
			buf.WriteByte(',')
		}
		buf.WriteByte('\n')
	}
	buf.WriteString("  ]")
	// Sort non-phases keys: Go map iteration is randomized, so sorting keeps
	// untouched top-level keys in a stable order across writes (no diff churn).
	for _, k := range sortedKeys(d.rest) {
		key, _ := json.Marshal(k)
		out, err := indentValue(d.rest[k], "  ")
		if err != nil {
			return nil, err
		}
		buf.WriteString(",\n  ")
		buf.Write(key)
		buf.WriteString(": ")
		buf.WriteString(out)
	}
	buf.WriteString("\n}\n")
	return buf.Bytes(), nil
}

// indentValue re-indents raw JSON so its first line sits flush (the caller
// writes the leading prefix) and continuation lines carry the given prefix.
func indentValue(raw json.RawMessage, prefix string) (string, error) {
	var compact bytes.Buffer
	if err := json.Compact(&compact, raw); err != nil {
		return "", fmt.Errorf("invalid json region: %w", err)
	}
	var out bytes.Buffer
	if err := json.Indent(&out, compact.Bytes(), prefix, "  "); err != nil {
		return "", err
	}
	return out.String(), nil
}

// backlogPhaseIndex returns the index of the Backlog phase, or -1.
func (d *rawDoc) backlogPhaseIndex() (int, error) {
	for i := range d.phases {
		name, err := phaseName(d.phaseBytes(i))
		if err != nil {
			return -1, err
		}
		if isBacklogPhaseName(name) {
			return i, nil
		}
	}
	return -1, nil
}

func (d *rawDoc) phaseBytes(i int) json.RawMessage {
	if dirty, ok := d.dirty[i]; ok {
		return json.RawMessage(dirty)
	}
	return d.phases[i]
}

func phaseName(p json.RawMessage) (string, error) {
	var obj struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(p, &obj); err != nil {
		return "", fmt.Errorf("invalid phase entry: %w", err)
	}
	return obj.Name, nil
}
