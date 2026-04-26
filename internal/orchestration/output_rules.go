package orchestration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func validateActionableOutputs(path, feature string, role Role, outputs []string) error {
	if role == RoleDocsSpecialist {
		return nil
	}
	files, missing := existingOutputFiles(outputs), missingOutputFiles(outputs)
	if len(missing) > 0 {
		return fmt.Errorf("actionable outputs must be real files; missing: %s in: %s", strings.Join(missing, ", "), path)
	}
	switch role {
	case RoleBigThinker, RoleFeatureSpecial:
		if hasPathPrefix(files, "docs/plans/") || hasPathPrefix(files, "specs/") {
			return nil
		}
		return fmt.Errorf("%s outputs must include a real plan or spec artifact in: %s", role, path)
	case RoleSeniorEngineer:
		if hasImplementationOutput(files) {
			return nil
		}
		return fmt.Errorf("senior-engineer outputs must include a real non-evidence implementation file in: %s", path)
	case RoleQASeniorEngineer:
		report := fmt.Sprintf(".workflow/%s-edge-cases.md", feature)
		if hasPathPrefix(files, "tests/") && containsPath(files, report) {
			return nil
		}
		return fmt.Errorf("qa-senior outputs must include at least one real test file and %s in: %s", report, path)
	default:
		return nil
	}
}

func existingOutputFiles(outputs []string) []string {
	files := []string{}
	for _, out := range outputs {
		path := normalizeOutputPath(out)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			files = append(files, path)
		}
	}
	return files
}

func missingOutputFiles(outputs []string) []string {
	missing := []string{}
	for _, out := range outputs {
		path := normalizeOutputPath(out)
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			missing = append(missing, out)
		}
	}
	return missing
}

func normalizeOutputPath(path string) string {
	clean := strings.TrimSpace(strings.TrimPrefix(path, "./"))
	return filepath.ToSlash(filepath.Clean(clean))
}

func hasPathPrefix(paths []string, prefix string) bool {
	for _, path := range paths {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func containsPath(paths []string, want string) bool {
	for _, path := range paths {
		if path == want {
			return true
		}
	}
	return false
}

func hasImplementationOutput(paths []string) bool {
	for _, path := range paths {
		if !hasPathPrefix([]string{path}, ".workflow/") && !hasPathPrefix([]string{path}, "tests/") &&
			!hasPathPrefix([]string{path}, "docs/features/") && !hasPathPrefix([]string{path}, "docs/plans/") &&
			!hasPathPrefix([]string{path}, "specs/") && !hasPathPrefix([]string{path}, "docs/project-docs/") {
			return true
		}
	}
	return false
}
