package setup

func planHooksSettings() (*SyncItem, error) {
	path := ".claude/settings.json"
	changed, _, err := buildHookSettings(path)
	if err != nil || !changed {
		return nil, err
	}
	return &SyncItem{Kind: SyncKindPrewriteHook, Path: path, Action: classifyAction(path)}, nil
}

func planOpenCodeConfig(local *LocalProvider) (*SyncItem, error) {
	path := "opencode.json"
	changed, _, err := buildOpenCodeConfig(path, local)
	if err != nil || !changed {
		return nil, err
	}
	// Carry local on the item so the apply path (ApplySync → InjectOpenCodeConfig)
	// writes the same managed provider block this plan detected.
	return &SyncItem{Kind: SyncOpenCodeCfg, Path: path, Action: classifyAction(path), Local: local}, nil
}
