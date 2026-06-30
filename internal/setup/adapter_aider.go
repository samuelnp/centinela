package setup

// aiderAdapter wires Aider: advisory governance only. Aider has no blocking
// prewrite hook, so it declares prompt-context + rules-file but NOT
// blocks-writes. It reuses AGENTS.md as its rules surface and points Aider at
// it via the .aider.conf.yml read: key.
type aiderAdapter struct{}

func (aiderAdapter) Name() string { return "aider" }

func (aiderAdapter) Capabilities() []Capability {
	return []Capability{CapPromptContext, CapRulesFile}
}

func (aiderAdapter) PlanItems() ([]SyncItem, error) {
	agents, err := planAgentsFile()
	if err != nil {
		return nil, err
	}
	cfg, err := planAiderConfig()
	if err != nil {
		return nil, err
	}
	return itemSlice(agents, cfg), nil
}
