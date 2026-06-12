package roadmap

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// render serializes the doc back to bytes at 2-space indent. Each phase is
// re-indented with json.Indent, which preserves key order and every field
// (including ones the Go structs drop) while normalizing whitespace
// deterministically. Dirty phases (Backlog / promote target) use rebuilt bytes.
func (d *rawDoc) render() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("{\n  \"phases\": [\n")
	for i := range d.phases {
		var out string
		var err error
		if dirty, ok := d.dirty[i]; ok {
			out, err = renderDirtyPhase(json.RawMessage(dirty)) // one feature per line
		} else {
			out, err = indentValue(d.phases[i], "    ") // untouched: re-indent only
		}
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
	for k, v := range d.rest {
		key, _ := json.Marshal(k)
		out, err := indentValue(v, "  ")
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
