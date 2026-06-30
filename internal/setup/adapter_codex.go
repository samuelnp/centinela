package setup

// codexAdapter wires OpenAI Codex as a full first-class harness: a blocking
// prewrite hook (via apply_patch in .codex/config.toml), prompt context (the
// UserPromptSubmit chain), and the shared AGENTS.md rules surface.
type codexAdapter struct{}

func (codexAdapter) Name() string { return "codex" }

func (codexAdapter) Capabilities() []Capability {
	return []Capability{CapBlocksWrites, CapPromptContext, CapRulesFile}
}

func (codexAdapter) PlanItems() ([]SyncItem, error) {
	cfg, err := planCodexConfig()
	if err != nil {
		return nil, err
	}
	agents, err := planAgentsFile()
	if err != nil {
		return nil, err
	}
	return itemSlice(cfg, agents), nil
}
