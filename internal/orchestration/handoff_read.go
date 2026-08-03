package orchestration

import (
	"encoding/json"
	"os"
)

// ReadHandoffTo returns the handoffTo field of an evidence file.
//
// It exists so a caller holding workflow state this package cannot see (the
// chain derivation in internal/workflow) can read one field without a second
// copy of the evidence JSON parsing. A missing or unparseable file is an error
// the caller is expected to IGNORE rather than report: ValidateRoles already
// names those, and a second message for the same defect only crowds the remedy.
func ReadHandoffTo(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var e struct {
		HandoffTo string `json:"handoffTo"`
	}
	if err := json.Unmarshal(data, &e); err != nil {
		return "", err
	}
	return e.HandoffTo, nil
}
