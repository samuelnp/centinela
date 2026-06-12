package evidence

import (
	"fmt"
	"os"

	"github.com/samuelnp/centinela/internal/workflow"
)

// DefaultCompanionTemplate returns the empty-but-structured markdown body the
// init subcommand drops alongside a fresh JSON skeleton. Agents overwrite it
// with their narrative; the validator only enforces existence. Known roles get
// a per-role FILL-slot skeleton; unknown roles fall back to a one-liner.
func DefaultCompanionTemplate(feature string, role Role) string {
	if body, ok := companionSkeleton(feature, role); ok {
		return fmt.Sprintf("# %s — %s\n\n%s", feature, role, body)
	}
	return fmt.Sprintf("# %s — %s\n\n_Replace this with the role's narrative report._\n", feature, role)
}

// WriteCompanion writes the markdown body next to the JSON via the same
// atomic temp-rename protocol. Callers are expected to hold the role lock.
func WriteCompanion(feature string, role Role, body string) error {
	if err := os.MkdirAll(workflow.WorkflowDir, 0o755); err != nil {
		return fmt.Errorf("evidence companion mkdir: %w", err)
	}
	return writeBytesAtomic(companionPath(feature, role), []byte(body))
}

// ReadCompanion returns the on-disk markdown body, or empty string + nil
// error if the file does not exist.
func ReadCompanion(feature string, role Role) (string, error) {
	data, err := os.ReadFile(companionPath(feature, role))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("evidence companion read: %w", err)
	}
	return string(data), nil
}
