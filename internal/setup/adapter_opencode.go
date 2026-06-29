package setup

// openCodeAdapter wires OpenCode: a blocking prewrite plugin, prompt context,
// and the AGENTS.md rules surface, plus opencode.json config.
type openCodeAdapter struct{}

func (openCodeAdapter) Name() string { return "opencode" }

func (openCodeAdapter) Capabilities() []Capability {
	return []Capability{CapBlocksWrites, CapPromptContext, CapRulesFile}
}

func (openCodeAdapter) PlanItems() ([]SyncItem, error) {
	cfg, err := planOpenCodeConfig()
	if err != nil {
		return nil, err
	}
	plug, err := planPluginFile()
	if err != nil {
		return nil, err
	}
	agents, err := planAgentsFile()
	if err != nil {
		return nil, err
	}
	return itemSlice(cfg, plug, agents), nil
}
