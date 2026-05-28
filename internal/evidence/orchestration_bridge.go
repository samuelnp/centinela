package evidence

import (
	"github.com/samuelnp/centinela/internal/orchestration"
)

// orchValidateImpl is the production hook through to the orchestration
// validator. Kept as its own file so tests can swap orchValidate without
// touching the bridge.
func orchValidateImpl(path, feature, step string, role Role, uiPaths []string) error {
	return orchestration.ValidateEvidence(path, feature, step, role, uiPaths)
}
