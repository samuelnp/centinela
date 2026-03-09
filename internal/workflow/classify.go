package workflow

import "strings"

// FileType categorises a file path by which workflow step governs it.
type FileType string

const (
	TypePlan  FileType = "plan"
	TypeCode  FileType = "code"
	TypeTests FileType = "tests"
	TypeOther FileType = "other"
)

// ClassifyFile maps a file path to the workflow step that owns it.
func ClassifyFile(path string) FileType {
	switch {
	case containsAny(path, "/docs/plans/", "/specs/"):
		return TypePlan
	case containsAny(path, "/src/", "/app/"):
		return TypeCode
	case strings.Contains(path, "/tests/"):
		return TypeTests
	default:
		return TypeOther
	}
}

// IsAllowedInStep returns true if a file type may be written during the given step.
func IsAllowedInStep(fileType FileType, step string) bool {
	switch step {
	case "plan":
		return fileType == TypePlan
	case "code":
		return fileType == TypeCode || fileType == TypePlan
	case "tests":
		return fileType == TypeTests || fileType == TypeCode
	case "validate":
		return true
	}
	return false
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
