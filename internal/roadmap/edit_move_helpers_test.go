package roadmap

import (
	"bytes"
	"os"
	"testing"
)

// canonRoadmap writes body, then rewrites it in the canonical on-disk render form
// (all phases through json.Indent). Untouched phases only round-trip byte-identical
// from a file already in this canonical shape, so edit/move/reorder tests seed here.
func canonRoadmap(t *testing.T, body string) (string, []byte) {
	t.Helper()
	p := crudWrite(t, body)
	doc, err := readRawRoadmap(p)
	if err != nil {
		t.Fatalf("readRawRoadmap: %v", err)
	}
	b, err := doc.render()
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return p, b
}

// phaseSlice returns the on-disk substring for phaseName from its name marker to
// the first features-array close. Asserting this slice survives a mutation proves
// an untouched phase round-tripped byte-identically.
func phaseSlice(t *testing.T, data []byte, phaseName string) []byte {
	t.Helper()
	i := bytes.Index(data, []byte(`"name": "`+phaseName+`"`))
	if i < 0 {
		t.Fatalf("phase %q not found in %s", phaseName, data)
	}
	j := bytes.Index(data[i:], []byte("]"))
	if j < 0 {
		t.Fatalf("no features close for phase %q", phaseName)
	}
	return data[i : i+j+1]
}

// contains reports whether names holds an exact match for want.
func contains(names []string, want string) bool {
	for _, n := range names {
		if n == want {
			return true
		}
	}
	return false
}

// orderIn returns the feature-name order of phaseName in the on-disk doc at path.
func orderIn(t *testing.T, path, phaseName string) []string {
	t.Helper()
	doc, err := readRawRoadmap(path)
	if err != nil {
		t.Fatalf("readRawRoadmap: %v", err)
	}
	order, err := doc.phaseOrder()
	if err != nil {
		t.Fatalf("phaseOrder: %v", err)
	}
	for i := range doc.phases {
		p, err := doc.decodePhase(i)
		if err != nil {
			t.Fatal(err)
		}
		if p.Name == phaseName {
			return order[i]
		}
	}
	t.Fatalf("phase %q not found", phaseName)
	return nil
}
