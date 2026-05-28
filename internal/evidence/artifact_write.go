package evidence

import (
	"errors"
	"fmt"
	"os"

	"github.com/samuelnp/centinela/internal/workflow"
)

// ErrArtifactExists signals an attempt to overwrite an existing artifact
// without --force. The CLI maps this to a non-zero exit and a hint.
var ErrArtifactExists = errors.New("artifact already exists")

// WriteArtifact renders the template for kind/feature and writes each file
// atomically. It refuses to overwrite existing files unless force is true.
// Returns the list of paths actually written.
func WriteArtifact(feature string, kind ArtifactKind, force bool) ([]string, error) {
	paths, bodies, err := RenderTemplate(kind, feature)
	if err != nil {
		return nil, err
	}
	if !force {
		if existing := firstExisting(paths); existing != "" {
			return nil, fmt.Errorf("%w: %s (use --force to overwrite)", ErrArtifactExists, existing)
		}
	}
	if err := os.MkdirAll(workflow.WorkflowDir, 0o755); err != nil {
		return nil, fmt.Errorf("artifact mkdir %s: %w", workflow.WorkflowDir, err)
	}
	for i, path := range paths {
		if err := writeBytesAtomic(path, bodies[i]); err != nil {
			return nil, err
		}
	}
	return paths, nil
}

func firstExisting(paths []string) string {
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
