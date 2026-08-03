package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// stateFileMode is the mode the state file is created with and kept at. It is
// applied by an explicit Chmod, so — unlike the os.WriteFile it replaced — it
// IGNORES the process umask, deliberately: every hook and subprocess (sometimes
// under another uid, in shared checkouts and CI images) must read this file, so
// its readability is a property of Centinela, not of the operator's shell
// (under `umask 077` the old writer produced 0600 and those readers failed).
// It holds coordination metadata, never secrets. A read-only 0400 file is
// normalised to 0644 too: rename needs directory write, not file write.
const stateFileMode fs.FileMode = 0o644

// Load reads and parses a workflow file from disk. Only a genuinely missing
// state file reports absence; read and parse failures surface with the state
// file path so they are never mistaken for "not started".
//
// Load is PERMISSIVE about the schema version by design: a file from a newer
// Centinela loads (its version is preserved on the returned struct and refused
// later, by Save). A failing Load would empty ActiveWorkflows and make the
// prewrite hook block every governed write — see SchemaVersion.
//
// The version is probed from the RAW bytes BEFORE the unmarshal, so that
// permissiveness never depends on the rest of the document being modellable: a
// future file that changed a type or a shape degrades (see degradedWorkflow)
// instead of failing. A file that does NOT claim a future version and cannot be
// parsed is genuine corruption and still errors.
func Load(feature string) (*Workflow, error) {
	path := FilePath(feature)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("no workflow found for %q", feature)
		}
		return nil, fmt.Errorf("reading workflow file %s: %w", path, err)
	}
	version, readable := stateVersion(data)
	var wf Workflow
	if err := json.Unmarshal(data, &wf); err != nil {
		if degraded := degradedWorkflow(feature, data, version, readable); degraded != nil {
			return degraded, nil
		}
		return nil, fmt.Errorf("invalid workflow file %s: %w", path, err)
	}
	if wf.SchemaVersion == 0 {
		wf.SchemaVersion = legacyVersion // in memory only; the file is not rewritten
	}
	wf.loadedDigest = digestOf(data)
	return &wf, nil
}

// Save writes a workflow to disk atomically.
//
// It first performs ONE read of the target that feeds both guards: the on-disk
// schema version (a newer binary may have upgraded the file since our Load, so
// the in-memory value is not authoritative) and the compare-and-swap staleness
// check. A missing target is a first write ONLY for a never-loaded workflow;
// for a loaded one it was removed underneath us — see checkNotDeleted.
func Save(wf *Workflow) error {
	path := FilePath(wf.Feature)
	current, err := os.ReadFile(path)
	switch {
	case err == nil:
		v, readable := stateVersion(current)
		if !readable {
			return errUnreadableVersion(path)
		}
		if v > SchemaVersion {
			return errFutureVersion(path, v)
		}
		if err := checkNotStale(path, wf, current); err != nil {
			return err
		}
	case errors.Is(err, fs.ErrNotExist):
		if err := checkNotDeleted(path, wf); err != nil {
			return err
		}
	default:
		return fmt.Errorf("reading workflow file %s: %w", path, err)
	}
	wf.SchemaVersion = SchemaVersion
	data, err := json.MarshalIndent(wf, "", "  ")
	if err != nil {
		return err
	}
	if err := WriteFileAtomic(path, data, stateFileMode); err != nil {
		return err
	}
	// This process is now the last writer, so a second Save on the same struct
	// must not conflict with the bytes we just published.
	wf.loadedDigest = digestOf(data)
	return nil
}
