package setup

const codexConfigFile = ".codex/config.toml"
const codexConfigHeader = "# centinela:managed-version=" + setupDocVersion + " template=.codex/config.toml"

// codexConfigBody is the fully-managed Codex hooks config. Codex uses nested
// matcher groups: [[hooks.<Event>]] carries a matcher, and a nested
// [[hooks.<Event>.hooks]] array lists the commands. apply_patch is Codex's
// canonical file-write tool, so it is the prewrite/postwrite matcher. The
// UserPromptSubmit chain mirrors the OpenCode plugin's prompt chain order.
const codexConfigBody = `[[hooks.PreToolUse]]
matcher = "apply_patch"

[[hooks.PreToolUse.hooks]]
type = "command"
command = "centinela hook prewrite"

[[hooks.PostToolUse]]
matcher = "apply_patch"

[[hooks.PostToolUse.hooks]]
type = "command"
command = "centinela hook postwrite"

[[hooks.UserPromptSubmit]]

[[hooks.UserPromptSubmit.hooks]]
type = "command"
command = "centinela hook setup"

[[hooks.UserPromptSubmit.hooks]]
type = "command"
command = "centinela hook migrate"

[[hooks.UserPromptSubmit.hooks]]
type = "command"
command = "centinela hook autostart"

[[hooks.UserPromptSubmit.hooks]]
type = "command"
command = "centinela hook orchestration"

[[hooks.UserPromptSubmit.hooks]]
type = "command"
command = "centinela hook plan-advisor"

[[hooks.UserPromptSubmit.hooks]]
type = "command"
command = "centinela hook context"
`

// planCodexConfig plans the managed .codex/config.toml via the shared
// managed-marker seam: absent -> create, managed -> update, unmanaged ->
// manual-review (never clobbered).
func planCodexConfig() (*SyncItem, error) {
	return planManagedFile(codexConfigFile, codexConfigHeader+"\n"+codexConfigBody, codexConfigBody, SyncKindPrewriteHook)
}

func writeManagedCodexConfig(path string) error {
	return writeManaged(path, codexConfigHeader+"\n"+codexConfigBody)
}
