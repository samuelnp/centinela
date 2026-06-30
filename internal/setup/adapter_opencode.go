package setup

// openCodeAdapter wires OpenCode: a blocking prewrite plugin, prompt context,
// and the AGENTS.md rules surface, plus opencode.json config. The registry holds
// the zero value (local == nil); BuildSyncPlanWithLocal substitutes an adapter
// carrying the managed local provider when a local block is declared.
type openCodeAdapter struct{ local *LocalProvider }

func (openCodeAdapter) Name() string { return "opencode" }

func (openCodeAdapter) Capabilities() []Capability {
	return []Capability{CapBlocksWrites, CapPromptContext, CapRulesFile}
}

func (a openCodeAdapter) PlanItems() ([]SyncItem, error) {
	cfg, err := planOpenCodeConfig(a.local)
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
