package evidence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/samuelnp/centinela/internal/workflow"
)

// pathFor returns the canonical .workflow/<feature>-<role>.json path. All
// reads and writes funnel through this helper so the worktree convention
// stays in one place.
func pathFor(feature string, role Role) string {
	return filepath.Join(workflow.WorkflowDir, fmt.Sprintf("%s-%s.json", feature, role))
}

// companionPath returns the .workflow/<feature>-<role>.md path.
func companionPath(feature string, role Role) string {
	return filepath.Join(workflow.WorkflowDir, fmt.Sprintf("%s-%s.md", feature, role))
}

// Read parses the on-disk evidence file for (feature, role). Missing files
// surface as a typed error so callers can give init-suggesting fix hints.
func Read(feature string, role Role) (*RoleEvidence, error) {
	path := pathFor(feature, role)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &NotFoundError{Path: path, Feature: feature, Role: role}
		}
		return nil, fmt.Errorf("evidence read %s: %w", path, err)
	}
	var r RoleEvidence
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("evidence parse %s: %w", path, err)
	}
	return &r, nil
}

// NotFoundError signals a missing file. Cobra commands distinguish it so
// they can suggest `centinela evidence init …`.
type NotFoundError struct {
	Path    string
	Feature string
	Role    Role
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("evidence not found: %s", e.Path)
}

// IsNotFound reports whether err is (or wraps) a NotFoundError.
func IsNotFound(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}
