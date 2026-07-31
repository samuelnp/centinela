package gatereport

import (
	"encoding/json"
	"fmt"
)

// ValidateCommandsSchema is the ONE place that knows the shape of the
// `commands` array inside a centinela:verification block. Both sides that
// touch the array route through it — ParseVerification on the read path and
// Stamped on the write path — so a malformed record is refused where it is
// WRITTEN, not only where it is later read.
//
// Shape: [{"argv": [string, ...] (non-empty), "exitCode": int, "durationMs":
// int (optional)}]. Unknown keys are ignored: the block is a forward-
// compatible record, and a future field must not brick an older binary.
//
// An absent array is not a schema failure — "no commands recorded" is an
// ADMISSIBILITY judgment that belongs to Assess, which reports it with the
// remedy an operator can act on.
func ValidateCommandsSchema(raw json.RawMessage) error {
	if len(raw) == 0 {
		return nil
	}
	var entries []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &entries); err != nil {
		return fmt.Errorf("commands must be an array of objects: %w", err)
	}
	for i, entry := range entries {
		if err := validateCommandEntry(entry); err != nil {
			return fmt.Errorf("commands[%d]: %w", i, err)
		}
	}
	return nil
}

// validateCommandEntry requires exitCode to be PRESENT, not merely
// unmarshallable. An absent key decodes to 0 in Command, which reads as
// "passed" — the same accepts-what-it-should-reject class this gate closes,
// so the key is demanded rather than defaulted.
func validateCommandEntry(entry map[string]json.RawMessage) error {
	rawArgv, has := entry["argv"]
	if !has {
		return fmt.Errorf("argv is required")
	}
	var argv []string
	if err := json.Unmarshal(rawArgv, &argv); err != nil {
		return fmt.Errorf("argv must be an array of strings: %w", err)
	}
	if len(argv) == 0 {
		return fmt.Errorf("empty argv (a recorded command with no argv is not evidence)")
	}
	rawExit, has := entry["exitCode"]
	if !has {
		return fmt.Errorf("exitCode is required (an absent key would silently read as a pass)")
	}
	var exitCode int
	if err := json.Unmarshal(rawExit, &exitCode); err != nil {
		return fmt.Errorf("exitCode must be an integer: %w", err)
	}
	if rawMS, ok := entry["durationMs"]; ok {
		var durationMS int
		if err := json.Unmarshal(rawMS, &durationMS); err != nil {
			return fmt.Errorf("durationMs must be an integer: %w", err)
		}
	}
	return nil
}
