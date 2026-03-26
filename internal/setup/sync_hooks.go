package setup

func planHooksSettings() (*SyncItem, error) {
	path := ".claude/settings.json"
	changed, _, err := buildHookSettings(path)
	if err != nil || !changed {
		return nil, err
	}
	return &SyncItem{Kind: SyncClaudeHooks, Path: path, Action: classifyAction(path)}, nil
}

func planOpenCodeConfig() (*SyncItem, error) {
	path := "opencode.json"
	changed, _, err := buildOpenCodeConfig(path)
	if err != nil || !changed {
		return nil, err
	}
	return &SyncItem{Kind: SyncOpenCodeCfg, Path: path, Action: classifyAction(path)}, nil
}
