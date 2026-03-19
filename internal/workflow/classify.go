package workflow

import (
	"strings"

	"github.com/samuelnp/centinela/internal/config"
)

// FileType categorises a file path by which workflow step governs it.
type FileType string

const (
	TypeRoadmap FileType = "roadmap"
	TypePlan    FileType = "plan"
	TypeCode    FileType = "code"
	TypeTests   FileType = "tests"
	TypeOther   FileType = "other"
)

// defaultCodeDirs covers common source roots across popular stacks.
// Go: cmd/, internal/, pkg/
// Ruby: lib/
// Generic: src/, app/, backend/, frontend/
var defaultCodeDirs = []string{
	"/src/", "/app/",
	"/cmd/", "/internal/", "/pkg/",
	"/lib/",
	"/backend/", "/frontend/",
}

// ClassifyFile maps a file path to the workflow step that owns it.
// codeDirs from cfg override the default set when non-empty.
func ClassifyFile(path string, cfg *config.Config) FileType {
	dirs := cfg.Workflow.CodeDirs
	if len(dirs) == 0 {
		dirs = defaultCodeDirs
	}
	switch {
	case isRoadmapArtifact(path):
		return TypeRoadmap
	case containsAny(path, "/docs/plans/", "/specs/"):
		return TypePlan
	case containsAny(path, dirs...):
		return TypeCode
	case strings.Contains(path, "/tests/"):
		return TypeTests
	default:
		return TypeOther
	}
}

// isRoadmapArtifact returns true for files belonging to the roadmap phase.
func isRoadmapArtifact(path string) bool {
	return strings.Contains(path, "/docs/features/") ||
		strings.HasSuffix(path, "ROADMAP.md") ||
		strings.HasSuffix(path, "roadmap.json")
}

// IsAllowedInStep returns true if a file type may be written during the given step.
func IsAllowedInStep(fileType FileType, step string) bool {
	switch step {
	case "plan":
		return fileType == TypePlan || fileType == TypeRoadmap
	case "code":
		return fileType == TypeCode || fileType == TypePlan || fileType == TypeRoadmap
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
