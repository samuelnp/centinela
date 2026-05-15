package main

import (
	"encoding/json"
	"os"
)

// readStewardHandoff extracts the steward verdict from already-validated
// evidence. orchestration.ValidateEvidence has confirmed the schema, so a
// missing handoffTo here is treated as an invalid resolution.
func readStewardHandoff(path string) (string, error) {
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
