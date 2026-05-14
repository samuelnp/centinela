package orchestration

import (
	"fmt"
	"strings"
)

func validateActionableOutputs(path, feature string, role Role, outputs, uiPaths []string) error {
	if role == RoleDocsSpecialist {
		return nil
	}
	files, missing := existingOutputFiles(outputs), missingOutputFiles(outputs)
	if len(missing) > 0 {
		return fmt.Errorf("actionable outputs must be real files; missing: %s in: %s", strings.Join(missing, ", "), path)
	}
	return dispatchRoleOutputs(path, feature, role, files, uiPaths)
}

func dispatchRoleOutputs(path, feature string, role Role, files, uiPaths []string) error {
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
	case RoleUXUISpecialist:
		return validateUXOutputs(path, files, uiPaths)
	case RoleMergeSteward:
		report := fmt.Sprintf(".workflow/%s-merge-steward.md", feature)
		if containsPath(files, report) {
			return nil
		}
		return fmt.Errorf("merge-steward outputs must include %s in: %s", report, path)
	default:
		return nil
	}
}
