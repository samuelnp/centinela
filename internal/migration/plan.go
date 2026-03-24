package migration

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/samuelnp/centinela/internal/scaffold"
)

func BuildPlan(root string) (Plan, error) {
	paths, err := managedPaths()
	if err != nil {
		return Plan{}, err
	}
	var plan Plan
	for _, p := range paths {
		item, ok, err := buildItem(root, p)
		if err != nil {
			return Plan{}, err
		}
		if ok {
			plan.Items = append(plan.Items, item)
		}
	}
	return plan, nil
}

func buildItem(root, path string) (Item, bool, error) {
	tpl, err := scaffold.ReadAsset(path)
	if err != nil {
		return Item{}, false, fmt.Errorf("read template %s: %w", path, err)
	}
	target := WithHeader(string(tpl), path, CurrentDocVersion)
	full := filepath.Join(root, path)
	current, err := os.ReadFile(full)
	if err != nil {
		if os.IsNotExist(err) {
			return Item{Path: path, Action: ActionCreate, ToVersion: CurrentDocVersion, content: target}, true, nil
		}
		return Item{}, false, err
	}
	h, ok := ParseHeader(string(current))
	if ok && h.Version == CurrentDocVersion && h.Template == path {
		return Item{}, false, nil
	}
	from := "legacy"
	if ok {
		from = h.Version
	}
	m, kept, custom := mergeContent(string(current), target)
	return Item{
		Path: path, Action: ActionUpdate, FromVersion: from, ToVersion: CurrentDocVersion,
		PreservedKeepBlocks: kept, PreservedCustomSection: custom, content: m,
	}, true, nil
}
