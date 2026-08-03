package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// stateFileMode is the mode the state file is created with and kept at. Every
// hook reads this file, so it must stay group/world readable.
const stateFileMode fs.FileMode = 0o644

// Load reads and parses a workflow file from disk. Only a genuinely
// missing state file reports absence; read and parse failures surface
// with the state file path so they are never mistaken for "not started".
//
// Load is PERMISSIVE about the schema version by design: a file from a newer
// Centinela loads normally (its version is preserved on the returned struct and
// refused later, by Save). Making Load fail here would empty ActiveWorkflows
// and make the prewrite hook block every governed write — see SchemaVersion.
func Load(feature string) (*Workflow, error) {
	path := FilePath(feature)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("no workflow found for %q", feature)
		}
		return nil, fmt.Errorf("reading workflow file %s: %w", path, err)
	}
	var wf Workflow
	if err := json.Unmarshal(data, &wf); err != nil {
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
// schema version (the file may have been upgraded by a newer binary since our
// Load, so the in-memory value is not authoritative) and the compare-and-swap
// staleness check. A missing target is neither problem — it is the first write.
func Save(wf *Workflow) error {
	path := FilePath(wf.Feature)
	current, err := os.ReadFile(path)
	switch {
	case err == nil:
		if v := onDiskVersion(current); v > SchemaVersion {
			return errFutureVersion(path, v)
		}
		if err := checkNotStale(path, wf, current); err != nil {
			return err
		}
	case !errors.Is(err, fs.ErrNotExist):
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
