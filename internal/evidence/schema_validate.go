package evidence

import (
	"errors"
	"strings"
)

// FieldError pairs a contract field with the error message reported by the
// orchestration validator. Field is best-effort: it is parsed from the
// validator's message; callers must not depend on a fully-typed field name.
type FieldError struct {
	Field   string
	Message string
}

// Validate delegates to ValidateOnDisk so the underlying rules in
// internal/orchestration apply unchanged. The evidence must already exist on
// disk for actionable-output checks. Returns nil-len slice when valid.
func (r *RoleEvidence) Validate(path string, uiPaths []string) []FieldError {
	err := orchValidate(path, r.Feature, r.Step, Role(r.Role), uiPaths)
	if err == nil {
		return nil
	}
	return []FieldError{classifyError(err)}
}

// classifyError parses an orchestration error message into a FieldError with
// a best-effort field guess. The full message is always preserved.
func classifyError(err error) FieldError {
	msg := err.Error()
	field := guessField(msg)
	return FieldError{Field: field, Message: msg}
}

func guessField(msg string) string {
	patterns := []struct {
		needle string
		field  string
	}{
		{"missing feature-doc snapshot inputs", "inputs"},
		{"actionable outputs must be real files", "outputs"},
		{"outputs must include", "outputs"},
		{"edgeCases required", "edgeCases"},
		{"mobileFirst", "mobileFirst"},
		{"missing required ux edgeCases", "edgeCases"},
		{"incomplete evidence fields", "incomplete"},
		{"invalid generatedAt", "generatedAt"},
		{"mismatched evidence fields", "feature"},
	}
	low := strings.ToLower(msg)
	for _, p := range patterns {
		if strings.Contains(low, strings.ToLower(p.needle)) {
			return p.field
		}
	}
	return ""
}

// orchValidate is a seam so tests can swap the orchestration validator out.
var orchValidate = orchValidateImpl

// errNilEvidence is returned by callers when validation is requested on a
// nil RoleEvidence pointer.
var errNilEvidence = errors.New("evidence: nil RoleEvidence")
