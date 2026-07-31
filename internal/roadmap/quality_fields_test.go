package roadmap

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestJsonKind(t *testing.T) {
	cases := map[string]string{
		"":        "null",
		"   ":     "null",
		"null":    "null",
		"{}":      "object",
		"[1]":     "array",
		`"x"`:     "string",
		"true":    "boolean",
		"false":   "boolean",
		"9":       "number",
		"-1.5":    "number",
		" \n{ } ": "object",
	}
	for raw, want := range cases {
		if got := jsonKind(json.RawMessage(raw)); got != want {
			t.Fatalf("jsonKind(%q) = %q, want %q", raw, got, want)
		}
	}
}

// The non-type-error fallback still names the feature and the schema rather
// than borrowing the range message.
func TestScoreTypeError_FallbackNamesTheSchema(t *testing.T) {
	err := scoreTypeError("user", errors.New("unexpected end of JSON input"))
	if !strings.Contains(err.Error(), `feature "user"`) ||
		!strings.Contains(err.Error(), "acceptanceCriteria") {
		t.Fatalf("fallback must name the feature and the schema, got %v", err)
	}
	if strings.Contains(err.Error(), "between 1 and 10") {
		t.Fatalf("a malformed scores object is not a range fault: %v", err)
	}
}

// features is an array, but of the wrong element type: still a structural error.
func TestDecodeQualityFeatures_WrongElementTypeIsStructural(t *testing.T) {
	raw := json.RawMessage(`[1, 2]`)
	_, err := decodeQualityFeatures(&raw)
	if err == nil || !strings.Contains(err.Error(), `"features" is malformed`) {
		t.Fatalf("expected a structural features error, got %v", err)
	}
	if strings.Contains(err.Error(), "between 1 and 10") {
		t.Fatalf("a malformed features array is not a range fault: %v", err)
	}
}
