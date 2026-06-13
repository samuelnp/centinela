package roadmap

import (
	"bytes"
	"encoding/json"
	"os"
)

// writeArtifact renders a top-level JSON object (analysis/quality) at 2-space
// indent, emitting the "features" array one object per line, and writes it
// atomically. Non-"features" keys are preserved as raw bytes, re-indented.
func writeArtifact(path string, top map[string]json.RawMessage) error {
	var buf bytes.Buffer
	buf.WriteString("{")
	first := true
	emit := func(key string, raw json.RawMessage) error {
		if !first {
			buf.WriteByte(',')
		}
		first = false
		buf.WriteString("\n  ")
		k, _ := json.Marshal(key)
		buf.Write(k)
		buf.WriteString(": ")
		out, err := indentValue(raw, "  ")
		if err != nil {
			return err
		}
		buf.WriteString(out)
		return nil
	}
	// Emit non-"features" keys in sorted order: Go map iteration is randomized,
	// so sorting guarantees byte-identical output across runs (no diff churn).
	for _, k := range sortedKeys(top) {
		if k == "features" {
			continue
		}
		if err := emit(k, top[k]); err != nil {
			return err
		}
	}
	if raw, ok := top["features"]; ok {
		if !first {
			buf.WriteByte(',')
		}
		if err := writeFeatureArray(&buf, raw); err != nil {
			return err
		}
	}
	buf.WriteString("\n}\n")
	return writeAtomic(path, buf.Bytes())
}

// writeFeatureArray emits a "features" array one compact object per line for
// merge-friendly diffs on appends.
func writeFeatureArray(buf *bytes.Buffer, raw json.RawMessage) error {
	var feats []json.RawMessage
	if err := json.Unmarshal(raw, &feats); err != nil {
		return err
	}
	buf.WriteString("\n  \"features\": [")
	for i, f := range feats {
		buf.WriteString("\n    ")
		var c bytes.Buffer
		if err := json.Compact(&c, f); err != nil {
			return err
		}
		buf.Write(c.Bytes())
		if i < len(feats)-1 {
			buf.WriteByte(',')
		}
	}
	if len(feats) > 0 {
		buf.WriteString("\n  ")
	}
	buf.WriteString("]")
	return nil
}

// appendLine appends a single line (plus newline) to a markdown file, creating
// it if absent. Used for promotion provenance bullets.
func appendLine(path, line string) error {
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	var buf bytes.Buffer
	buf.Write(existing)
	if len(existing) > 0 && existing[len(existing)-1] != '\n' {
		buf.WriteByte('\n')
	}
	buf.WriteString(line)
	buf.WriteByte('\n')
	return writeAtomic(path, buf.Bytes())
}
