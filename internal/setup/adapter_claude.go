package setup

// claudeAdapter wires Claude Code: a blocking prewrite hook, prompt context,
// and a rules surface, all merged into .claude/settings.json.
type claudeAdapter struct{}

func (claudeAdapter) Name() string { return "claude" }

func (claudeAdapter) Capabilities() []Capability {
	return []Capability{CapBlocksWrites, CapPromptContext, CapRulesFile}
}

func (claudeAdapter) PlanItems() ([]SyncItem, error) {
	it, err := planHooksSettings()
	if err != nil {
		return nil, err
	}
	return itemSlice(it), nil
}
