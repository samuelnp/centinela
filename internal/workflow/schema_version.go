package workflow

import (
	"encoding/json"
	"fmt"
)

// SchemaVersion is the .workflow/<feature>.json schema this binary writes and
// understands. Every Save stamps it.
//
// Migration contract:
//
//   - ABSENT or 0 on disk means version 1. Every field added to Workflow so far
//     is back-compat-by-absence (Archetype, ValidateContract, PlanContract,
//     ModelRoutes, ...), so "absent means the documented default" is the whole
//     of that migration. Nothing is asked of the operator.
//   - EQUAL loads and round-trips with every field intact.
//   - LOWER loads and is silently upgraded by the next Save. A future version
//     that needs more than defaulting must add its migration HERE, keyed off
//     the on-disk version, before that upgrade is allowed to happen.
//   - HIGHER loads, and Save refuses (see errFutureVersion). Saving would drop
//     the fields this binary does not model.
//
// The refusal deliberately lives in Save and never in Load. A Load that failed
// on a future version would cascade: ActiveWorkflows warns and skips the
// feature, the active set goes empty, EvaluatePrewrite returns NeedInit, every
// governed write is blocked, and the autostart hook then starts a duplicate
// workflow. Refusing in Save costs the operator only the ability to ADVANCE
// that one feature, which is the correct blast radius.
//
// Limitation worth stating plainly: versioning protects forward only. Binaries
// released before this check have no refusal, so they can still silently drop
// fields from a file this release writes. The guarantee begins once both sides
// carry the check; nothing can retrofit it.
const SchemaVersion = 1

// legacyVersion is the version attributed to a state file written before the
// version field existed. Spelled separately from SchemaVersion so that bumping
// the constant does not silently re-label every legacy file.
const legacyVersion = 1

// versionProbe reads only the version key. Decoding into it tolerates every
// other field, including fields a newer binary wrote that this one cannot
// model — which is the whole point of probing rather than unmarshalling into
// Workflow.
type versionProbe struct {
	SchemaVersion int `json:"schemaVersion"`
}

// onDiskVersion reports the schema version recorded in raw. Unparseable bytes
// and an absent key both report the legacy version: neither is evidence of a
// file from the future, and deciding what to do about a corrupt state file is
// not this function's job.
func onDiskVersion(raw []byte) int {
	var p versionProbe
	if err := json.Unmarshal(raw, &p); err != nil || p.SchemaVersion == 0 {
		return legacyVersion
	}
	return p.SchemaVersion
}

// errFutureVersion is the refusal Save reports when the file on disk was
// written by a newer Centinela. It names the file, the version it carries, the
// version this binary understands, and the fix.
func errFutureVersion(path string, onDisk int) error {
	return fmt.Errorf("%s was written by a newer Centinela (schema version %d); "+
		"this binary understands schema version %d. Refusing to write — saving "+
		"would drop fields it does not know about. Upgrade with `centinela "+
		"update` and re-run", path, onDisk, SchemaVersion)
}
