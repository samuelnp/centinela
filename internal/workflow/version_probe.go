package workflow

import "encoding/json"

// stateFields decodes raw as a flat key→raw-token map, tolerating every value
// shape a newer binary might have written. nil means "not a JSON object", which
// is corruption, not evidence of a file from the future.
func stateFields(raw []byte) map[string]json.RawMessage {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil
	}
	return fields
}

// stateVersion reports the schema version recorded in raw and whether it could
// be read at all. It is deliberately independent of Workflow: probing a single
// raw token is what lets a file whose fields this binary cannot model still
// yield its version.
//
//	(1, true)      — absent, null or 0: a legacy file (see legacyVersion).
//	(n, true)      — an integer version; n > SchemaVersion is a future file.
//	(0, false)     — present but unreadable as an int ("2.0", 1.5, true,
//	                 99999999999999999999). Treated exactly like a future
//	                 version: this binary demonstrably does not understand the
//	                 file's schema.
//	(1, true)      — raw is not a JSON object at all. Corruption is not evidence
//	                 of the future; diagnosing it is not this function's job.
func stateVersion(raw []byte) (int, bool) {
	fields := stateFields(raw)
	if fields == nil {
		return legacyVersion, true
	}
	token, present := fields["schemaVersion"]
	if !present || string(token) == "null" {
		return legacyVersion, true
	}
	var v int
	if err := json.Unmarshal(token, &v); err != nil {
		return 0, false
	}
	if v == 0 {
		return legacyVersion, true
	}
	return v, true
}
