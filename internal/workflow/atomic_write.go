package workflow

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// beforeRename runs after the temp file is fully written and fsynced but
// before the rename publishes it. It is a no-op in production and exists so a
// test can die in exactly that window (os.Exit, which skips deferred cleanup
// the way a SIGKILL does) and prove the target survives intact.
var beforeRename = func() {}

// WriteFileAtomic replaces path with data in one observable step: a uniquely
// named sibling temp file is written, chmod'd to perm, fsynced, then renamed
// over the target. A reader therefore sees either the complete old file or the
// complete new one, never the zero-byte window os.WriteFile's O_TRUNC opens.
//
// The temp lives beside the target, so rename(2) stays within one filesystem
// and cannot fail with EXDEV. Its name — ".<base>.tmp-<rand>" — is load-bearing
// twice over. The random suffix keeps two concurrent writers off each other's
// temp file (a shared name would let one rename the other's half-written bytes
// over the target). The dot prefix and the "-<rand>" tail keep it out of both
// ".workflow/*.json" (ActiveWorkflows) and ".workflow/*.json.tmp" (the doctor
// evidence check, whose repair only knows how to remove <feature>-<role> temps
// and would otherwise report a fix that removes nothing, forever).
//
// SYMLINKS ARE REPLACED, NOT FOLLOWED — accepted deliberately. rename(2)
// replaces the link itself, where os.WriteFile wrote through it. Writing
// through a link would re-open the O_TRUNC window this helper exists to close,
// and the link's target may sit on another filesystem where rename cannot be
// atomic at all (EXDEV). Share state by symlinking the .workflow/ DIRECTORY:
// that still works, since the temp lands in the resolved directory.
//
// Every error names the TARGET, not the temp: the caller asked to write the
// state file and that is the path worth reporting.
func WriteFileAtomic(path string, data []byte, perm fs.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := writeTempSibling(dir, filepath.Base(path), data, perm)
	if err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	// A no-op once the rename has moved the file; the real work is cleaning up
	// after a failed rename.
	defer os.Remove(tmp) //nolint:errcheck // best-effort cleanup
	beforeRename()
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	syncDir(dir)
	return nil
}

// writeTempSibling creates ".<base>.tmp-<rand>" in dir, writes data, applies
// perm, fsyncs, and closes, removing the partial file on any failure.
//
// The explicit Chmod is not decoration: os.CreateTemp opens at 0600, so
// without it every save would silently downgrade the state file that every
// hook reads to owner-only. Write/chmod/sync/close errors are coalesced —
// the caller only needs to know the write failed, not which syscall said so.
func writeTempSibling(dir, base string, data []byte, perm fs.FileMode) (string, error) {
	f, err := os.CreateTemp(dir, "."+base+".tmp-")
	if err != nil {
		return "", err
	}
	tmp := f.Name()
	_, werr := f.Write(data)
	perr := f.Chmod(perm)
	serr := f.Sync()
	cerr := f.Close()
	if err := firstNonNil(werr, perr, serr, cerr); err != nil {
		_ = os.Remove(tmp)
		return "", err
	}
	return tmp, nil
}

// syncDir fsyncs dir so the rename's directory entry survives a power loss.
// Best-effort by design: opening a directory for sync is not portable, and a
// failure here must never fail a save whose rename already succeeded.
func syncDir(dir string) {
	d, err := os.Open(dir)
	if err != nil {
		return
	}
	_ = d.Sync()
	_ = d.Close()
}

func firstNonNil(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}
