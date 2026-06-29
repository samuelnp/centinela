package setup

import "os"

// BuildSyncPlan composes the managed-asset plan for an agent selector by
// iterating the harness registry. There is no per-harness if-ladder: single
// agents and composites (e.g. "both") both resolve through adaptersFor.
func BuildSyncPlan(agent string) (SyncPlan, error) {
	adapters, err := adaptersFor(agent)
	if err != nil {
		return SyncPlan{}, err
	}
	plan := SyncPlan{}
	for _, a := range adapters {
		items, err := a.PlanItems()
		if err != nil {
			return SyncPlan{}, err
		}
		plan.Items = append(plan.Items, items...)
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
	case SyncKindPrewriteHook:
		if it.Path == pluginFile {
			return writeManagedPlugin(it.Path)
		}
		_, err := InjectHooks(it.Path)
		return err
	case SyncOpenCodeCfg:
		_, err := InjectOpenCodeConfig(it.Path)
		return err
	case SyncAgents:
		return writeManagedAgents(it.Path)
	case SyncAiderConfig:
		return writeManagedAiderConfig(it.Path)
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

// itemSlice collects non-nil planned items into a slice.
func itemSlice(items ...*SyncItem) []SyncItem {
	out := []SyncItem{}
	for _, it := range items {
		if it != nil {
			out = append(out, *it)
		}
	}
	return out
}

func classifyAction(path string) SyncAction {
	if _, err := os.Stat(path); err == nil {
		return SyncUpdate
	}
	return SyncCreate
}
