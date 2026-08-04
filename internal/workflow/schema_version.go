package workflow

import "fmt"

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
//   - HIGHER, or UNREADABLE (see stateVersion), loads and Save refuses. Saving
//     would drop the fields this binary does not model.
//
// The refusal deliberately lives in Save and never in Load. A Load that failed
// on a future version would cascade: ActiveWorkflows warns and skips the
// feature, the active set goes empty, EvaluatePrewrite returns NeedInit, every
// governed write is blocked, and the autostart hook then starts a duplicate
// workflow. Refusing in Save costs the operator only the ability to ADVANCE
// that one feature, which is the correct blast radius.
//
// That guarantee must not depend on the REST of the document being modellable,
// or it holds only for purely additive future versions — a constraint nothing
// could enforce. The version is therefore probed from the raw bytes BEFORE any
// full unmarshal, and a future-versioned file this binary cannot unmarshal
// still loads, degraded. The full Load contract is in state_future.go.
//
// Adding a field to Workflow without bumping this constant is a silent
// data-loss release: an equal-version Save re-marshals from the struct and
// drops every on-disk key the struct does not model.
// TestWorkflowStructFieldsAreVersionLocked pins the field list for exactly that
// reason.
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

// errFutureVersion is the refusal Save reports when the file on disk was
// written by a newer Centinela. It names the file, the version it carries, the
// version this binary understands, and the fix.
func errFutureVersion(path string, onDisk int) error {
	return fmt.Errorf("%s was written by a newer Centinela (schema version %d); "+
		"this binary understands schema version %d. Refusing to write — saving "+
		"would drop fields it does not know about. Upgrade with `centinela "+
		"update` and re-run", path, onDisk, SchemaVersion)
}

// errUnreadableVersion is the refusal Save reports when the file carries a
// schemaVersion that is not an integer this binary can read (a string, a float,
// a bool, an out-of-range number). It fails CLOSED on purpose: a guard whose
// whole job is refusing to write what it cannot understand must not read "I
// could not parse this file's version" as "version 1". Only an ABSENT key is
// legacy.
func errUnreadableVersion(path string) error {
	return fmt.Errorf("%s carries a schemaVersion this binary cannot read, so it "+
		"was written by a newer Centinela; this binary understands schema "+
		"version %d. Refusing to write — saving would drop fields it does not "+
		"know about. Upgrade with `centinela update` and re-run",
		path, SchemaVersion)
}
