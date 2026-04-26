package orchestration

import "fmt"

func validateUXOutputs(path string, files, uiPaths []string) error {
	if hasAnyPrefix(files, uiPaths) {
		return nil
	}
	return fmt.Errorf("ux-ui-specialist outputs must include at least one real UI file in: %s", path)
}

func hasAnyPrefix(paths, prefixes []string) bool {
	for _, prefix := range prefixes {
		if hasPathPrefix(paths, prefix) {
			return true
		}
	}
	return false
}
