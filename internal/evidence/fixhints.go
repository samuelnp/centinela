package evidence

import (
	"fmt"
	"strings"
)

// FixHint pairs a validation error with the exact CLI command an agent can
// run to fix it. Each hint targets one (feature, role, field) tuple.
type FixHint struct {
	Feature string
	Role    Role
	Field   string
	Message string
	Command string
}

// String renders the hint as one human-readable line for stderr.
func (h FixHint) String() string {
	if h.Command == "" {
		return fmt.Sprintf("[%s/%s] %s", h.Feature, h.Role, h.Message)
	}
	return fmt.Sprintf("[%s/%s] %s\n  fix: %s", h.Feature, h.Role, h.Message, h.Command)
}

// fixHintFor renders a typed FixHint from a validator FieldError plus the
// (feature, role) context. The Command field maps to the smallest CLI
// invocation that closes the loop — append for lists, set for scalars,
// set extra.<k> for free-form.
func fixHintFor(feature string, role Role, fe FieldError) FixHint {
	cmd := suggestCommand(feature, role, fe)
	return FixHint{
		Feature: feature, Role: role, Field: fe.Field,
		Message: fe.Message, Command: cmd,
	}
}

func suggestCommand(feature string, role Role, fe FieldError) string {
	switch fe.Field {
	case "inputs", "outputs", "edgeCases":
		return fmt.Sprintf("centinela evidence append %s %s %s <value>", feature, role, fe.Field)
	case "mobileFirst":
		return fmt.Sprintf("centinela evidence set %s %s mobileFirst true", feature, role)
	case "status":
		return fmt.Sprintf("centinela evidence set %s %s status done", feature, role)
	case "generatedAt":
		return fmt.Sprintf("centinela evidence set %s %s generatedAt <RFC3339-timestamp>", feature, role)
	case "feature", "step", "role":
		return fmt.Sprintf("centinela evidence init %s %s  # re-create with correct context", feature, role)
	case "incomplete":
		return fmt.Sprintf("centinela evidence append %s %s inputs <path>  # and append outputs / set handoffTo as needed", feature, role)
	}
	if strings.Contains(strings.ToLower(fe.Message), "missing evidence") {
		return fmt.Sprintf("centinela evidence init %s %s", feature, role)
	}
	return ""
}
