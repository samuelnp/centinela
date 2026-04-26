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
	HandoffTo   string   `json:"handoffTo"`
	Checksum    string   `json:"checksum,omitempty"`
}

func ValidateEvidence(path, feature, step string, role Role) error {
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
	if (role == RoleFeatureSpecial || role == RoleQASeniorEngineer) && len(e.EdgeCases) == 0 {
		return fmt.Errorf("edgeCases required in: %s", path)
	}
	if _, err := time.Parse(time.RFC3339, strings.TrimSpace(e.GeneratedAt)); err != nil {
		return fmt.Errorf("invalid generatedAt in: %s", path)
	}
	if err := validateActionableOutputs(path, feature, role, e.Outputs); err != nil {
		return err
	}
	if err := validatePlanSnapshotInputs(path, feature, step, role, e.Inputs); err != nil {
		return err
	}
	return nil
}
