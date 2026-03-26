package setup

import "os"

func BuildSyncPlan(agent string) (SyncPlan, error) {
	plan := SyncPlan{}
	if useClaude(agent) {
		it, err := planHooksSettings()
		if err != nil {
			return SyncPlan{}, err
		}
		appendItem(&plan, it)
	}
	if useOpenCode(agent) {
		it, err := planOpenCodeConfig()
		if err != nil {
			return SyncPlan{}, err
		}
		appendItem(&plan, it)
		it, err = planPluginFile()
		if err != nil {
			return SyncPlan{}, err
		}
		appendItem(&plan, it)
		it, err = planAgentsFile()
		if err != nil {
			return SyncPlan{}, err
		}
		appendItem(&plan, it)
	}
	return plan, nil
}

func ApplySync(plan SyncPlan) error {
	for _, it := range plan.Items {
		if it.Action == SyncManualReview {
			continue
		}
		if err := applyItem(it); err != nil {
			return err
		}
	}
	return nil
}

func applyItem(it SyncItem) error {
	switch it.Kind {
	case SyncClaudeHooks:
		_, err := InjectHooks(it.Path)
		return err
	case SyncOpenCodeCfg:
		_, err := InjectOpenCodeConfig(it.Path)
		return err
	case SyncOpenCodePlug:
		return writeManagedPlugin(it.Path)
	case SyncAgents:
		return writeManagedAgents(it.Path)
	default:
		return nil
	}
}

func appendItem(plan *SyncPlan, item *SyncItem) {
	if item == nil {
		return
	}
	plan.Items = append(plan.Items, *item)
}

func useClaude(agent string) bool {
	return agent == "claude" || agent == "both"
}

func useOpenCode(agent string) bool {
	return agent == "opencode" || agent == "both"
}

func classifyAction(path string) SyncAction {
	if _, err := os.Stat(path); err == nil {
		return SyncUpdate
	}
	return SyncCreate
}
