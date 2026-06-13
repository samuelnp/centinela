package roadmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// rawDoc is roadmap.json parsed with phases kept addressable while every other
// top-level key and every untouched phase is preserved as raw bytes.
type rawDoc struct {
	phases []json.RawMessage
	rest   map[string]json.RawMessage
	dirty  map[int]string // phase index -> rebuilt JSON bytes
}

// readRawRoadmap loads roadmap.json preserving unknown fields and formatting of
// untouched regions.
func readRawRoadmap(path string) (*rawDoc, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return nil, fmt.Errorf("invalid roadmap json: %w", err)
	}
	doc := &rawDoc{rest: map[string]json.RawMessage{}, dirty: map[int]string{}}
	for k, v := range top {
		if k == "phases" {
			if err := json.Unmarshal(v, &doc.phases); err != nil {
				return nil, fmt.Errorf("invalid roadmap phases: %w", err)
			}
			continue
		}
		doc.rest[k] = v
	}
	return doc, nil
}

// writeRawRoadmap renders the doc and writes it via temp-file+rename. The
// Backlog phase's features array is emitted one object per line so concurrent
// appends conflict as a trivial textual union.
func writeRawRoadmap(path string, doc *rawDoc) error {
	body, err := doc.render()
	if err != nil {
		return err
	}
	return writeAtomic(path, body)
}

// writeAtomic writes data to path via a sibling temp file then renames it.
func writeAtomic(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}

// compactBytes returns a stable single-line JSON encoding of v.
func compactBytes(v any) (json.RawMessage, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return json.RawMessage(bytes.TrimRight(buf.Bytes(), "\n")), nil
}
