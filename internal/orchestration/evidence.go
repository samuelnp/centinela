package orchestration

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type Evidence struct {
	Feature     string   `json:"feature"`
	Step        string   `json:"step"`
	Role        string   `json:"role"`
	Status      string   `json:"status"`
	GeneratedAt string   `json:"generatedAt"`
	Inputs      []string `json:"inputs"`
	Outputs     []string `json:"outputs"`
	EdgeCases   []string `json:"edgeCases"`
	MobileFirst *bool    `json:"mobileFirst,omitempty"`
	HandoffTo   string   `json:"handoffTo"`
	Checksum    string   `json:"checksum,omitempty"`
}

func ValidateEvidence(path, feature, step string, role Role, uiPaths []string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("missing evidence json: %s", path)
	}
	var e Evidence
	if err := json.Unmarshal(data, &e); err != nil {
		return fmt.Errorf("invalid evidence json: %s", path)
	}
	if e.Feature != feature || e.Step != step || e.Role != string(role) {
		return fmt.Errorf("mismatched evidence fields: %s", path)
	}
	if e.Status != "done" || len(e.Inputs) == 0 || len(e.Outputs) == 0 || e.HandoffTo == "" {
		return fmt.Errorf("incomplete evidence fields: %s", path)
	}
	if needsEdgeCases(role) && len(e.EdgeCases) == 0 {
		return fmt.Errorf("edgeCases required in: %s", path)
	}
	if _, err := time.Parse(time.RFC3339, strings.TrimSpace(e.GeneratedAt)); err != nil {
		return fmt.Errorf("invalid generatedAt in: %s", path)
	}
	if err := validateActionableOutputs(path, feature, role, e.Outputs, uiPaths); err != nil {
		return err
	}
	if err := validateUXEvidence(path, role, e.EdgeCases, e.MobileFirst); err != nil {
		return err
	}
	if err := validatePlanSnapshotInputs(path, feature, step, role, e.Inputs); err != nil {
		return err
	}
	return validateStewardHandoff(path, role, e.HandoffTo)
}

// validateStewardHandoff binds the out-of-band merge-steward role's handoff to
// its two legal verdicts. The steward runs on `centinela merge`, never appears
// in a workflow's OrderedSteps(), and so has no derivable successor — its rule
// is a closed literal pair (APPLY -> "complete", ESCALATE -> "user"), which is
// exactly what worktree finalization already branches on. Any other value
// silently reads as ESCALATE there, stalling a merge with no stated reason.
//
// Every OTHER role's handoffTo is checked against the chain its own workflow
// derives, which needs workflow state this package cannot see; that check
// lives in internal/workflow.
func validateStewardHandoff(path string, role Role, handoffTo string) error {
	if role != RoleMergeSteward || handoffTo == "complete" || handoffTo == "user" {
		return nil
	}
	return fmt.Errorf("merge-steward handoffTo must be \"complete\" (APPLY) or \"user\" (ESCALATE), got %q: %s", handoffTo, path)
}

// needsEdgeCases includes RolePlanner because the planner now carries the spec
// lens that used to belong to the feature-specialist.
func needsEdgeCases(role Role) bool {
	return role == RolePlanner || role == RoleFeatureSpecial ||
		role == RoleQASeniorEngineer || role == RoleUXUISpecialist
}
