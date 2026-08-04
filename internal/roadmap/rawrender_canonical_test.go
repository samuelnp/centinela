package roadmap

import (
	"strings"
	"testing"
)

const canonBody = `{"intro":"hi","phases":[` +
	`{"name":"Phase 1","note":"n","features":[{"name":"a","description":"first"},{"name":"b"}]},` +
	`{"name":"Phase 2","features":[{"name":"c"}]},` +
	`{"name":"Backlog","features":[{"name":"f","summary":"s"}]}]}`

// AC14: EVERY phase renders one compact feature object per line, not only the
// phase the mutation touched.
func TestRenderIsCanonicalForEveryPhase(t *testing.T) {
	out := renderStr(t, docFrom(t, canonBody))
	for _, want := range []string{
		`{"name":"a","description":"first"}`, `{"name":"b"}`,
		`{"name":"c"}`, `{"name":"f","summary":"s"}`,
	} {
		if !strings.Contains(out, "\n        "+want) {
			t.Fatalf("%s is not on its own line:\n%s", want, out)
		}
	}
}

// Idempotence is what makes the format a fixed point: without it every write
// would churn the whole file.
func TestRenderIsIdempotent(t *testing.T) {
	p := crudWrite(t, canonBody)
	doc, err := readRawRoadmap(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeRawRoadmap(p, doc); err != nil {
		t.Fatal(err)
	}
	first := string(crudBytes(t, p))
	reread, err := readRawRoadmap(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeRawRoadmap(p, reread); err != nil {
		t.Fatal(err)
	}
	if second := string(crudBytes(t, p)); second != first {
		t.Fatalf("a second write must be byte-identical:\n%q\n%q", first, second)
	}
}

// A one-field edit must move exactly one line — the diff-churn kill.
func TestCanonicalEditTouchesOneLine(t *testing.T) {
	p := crudWrite(t, canonBody)
	doc, err := readRawRoadmap(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeRawRoadmap(p, doc); err != nil {
		t.Fatal(err)
	}
	before := strings.Split(string(crudBytes(t, p)), "\n")
	if err := Edit(p, EditRequest{Slug: "a", Description: "second"}); err != nil {
		t.Fatal(err)
	}
	after := strings.Split(string(crudBytes(t, p)), "\n")
	if len(before) != len(after) {
		t.Fatalf("line count changed: %d -> %d", len(before), len(after))
	}
	differing := 0
	for i := range before {
		if before[i] != after[i] {
			differing++
		}
	}
	if differing != 1 {
		t.Fatalf("a one-field edit changed %d lines, want 1", differing)
	}
}
