package evidence

import (
	"strings"
	"testing"
)

func TestMarshalSkipsNilOptionalFields(t *testing.T) {
	r := &RoleEvidence{Feature: "alpha", Role: "x", Step: "y", Status: "done", HandoffTo: "z"}
	out, err := r.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), `"_meta"`) {
		t.Fatalf("nil _meta should not appear: %s", out)
	}
	if strings.Contains(string(out), `"mobileFirst"`) {
		t.Fatalf("nil mobileFirst should not appear: %s", out)
	}
}

func TestMarshalNoEscapeKeepsAngleBrackets(t *testing.T) {
	out, err := marshalNoEscape("<feature>")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "<feature>") {
		t.Fatalf("angle brackets escaped: %s", out)
	}
}

func TestNonNilStringsReplacesNil(t *testing.T) {
	if got := nonNilStrings(nil); len(got) != 0 {
		t.Fatalf("expected empty slice, got %v", got)
	}
	in := []string{"a"}
	if got := nonNilStrings(in); &got[0] != &in[0] {
		t.Fatal("non-nil slice should not be reallocated")
	}
}

func TestPrettyIndentInvalidInput(t *testing.T) {
	if _, err := prettyIndent([]byte("not json")); err == nil {
		t.Fatal("expected error on invalid json")
	}
}

func TestUnmarshalJSONRejectsGarbage(t *testing.T) {
	var r RoleEvidence
	if err := r.UnmarshalJSON([]byte("not json")); err == nil {
		t.Fatal("expected error")
	}
}

func TestUnmarshalJSONTypeMismatch(t *testing.T) {
	raw := []byte(`{"feature": 42}`)
	var r RoleEvidence
	if err := r.UnmarshalJSON(raw); err == nil {
		t.Fatal("expected type error for non-string feature")
	}
}

func TestMarshalIncludesMobileFirstWhenSet(t *testing.T) {
	tr := true
	r := &RoleEvidence{Feature: "alpha", Role: "ux-ui-specialist", Step: "code", Status: "done", HandoffTo: "qa-senior", MobileFirst: &tr}
	out, err := r.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `"mobileFirst": true`) {
		t.Fatalf("missing mobileFirst: %s", out)
	}
}
