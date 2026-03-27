package setup

import "os"

const agentsFile = "AGENTS.md"

const agentsContent = `# Centinela Rules

This project uses Centinela workflow enforcement.

## Mandatory
- Read and follow CLAUDE.md for framework rules.
- Read PROJECT.md before planning or coding.
- Start every feature with centinela start <feature>.
- Do not bypass workflow order: plan -> code -> tests -> validate.

## OpenCode Integration
- Centinela prewrite checks are enforced by .opencode/plugins/centinela.js.
- If a write is blocked, fix step alignment instead of forcing the write.

## Commands
- centinela start <feature>
- centinela complete <feature>
- centinela status <feature>
- centinela validate
- centinela docs validate
- centinela docs generate --out docs/project-docs/index.html
`

// EnsureAgentsFile writes AGENTS.md for OpenCode if missing.
func EnsureAgentsFile() (bool, error) {
	if _, err := os.Stat(agentsFile); err == nil {
		return false, nil
	}
	return true, os.WriteFile(agentsFile, []byte(agentsContent), 0644)
}
