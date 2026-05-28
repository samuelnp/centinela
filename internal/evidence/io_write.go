package evidence

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/samuelnp/centinela/internal/workflow"
)

// tempSuffix is appended to evidence paths during atomic writes. Kept
// constant so `centinela evidence repair` can deterministically find and
// remove orphans after a crash mid-write.
const tempSuffix = ".tmp"

// WriteAtomic writes the marshalled RoleEvidence to disk via temp-file +
// fsync + rename. The temp file is sibling-named so the rename stays on the
// same filesystem (no EXDEV).
func WriteAtomic(feature string, role Role, r *RoleEvidence) error {
	if r == nil {
		return errNilEvidence
	}
	if err := os.MkdirAll(workflow.WorkflowDir, 0o755); err != nil {
		return fmt.Errorf("evidence mkdir %s: %w", workflow.WorkflowDir, err)
	}
	data, err := r.MarshalJSON()
	if err != nil {
		return fmt.Errorf("evidence marshal: %w", err)
	}
	return writeBytesAtomic(pathFor(feature, role), data)
}

// writeBytesAtomic implements the temp-write + fsync + rename ritual. All
// errors are wrapped with the target path so the agent sees what failed.
func writeBytesAtomic(target string, data []byte) error {
	tmp := target + tempSuffix
	if err := writeTempFile(tmp, data); err != nil {
		return err
	}
	if err := os.Rename(tmp, target); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("evidence rename %s -> %s: %w", tmp, target, err)
	}
	return nil
}

// writeTempFile opens, writes, fsyncs, and closes the temp file in one go,
// removing the partial file on any failure. Errors from a temp file we
// just opened (write/sync/close) are coalesced — the agent only cares
// that the write failed, not which call surfaced the syscall error.
func writeTempFile(tmp string, data []byte) error {
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("evidence open temp %s: %w", tmp, err)
	}
	_, werr := f.Write(data)
	serr := f.Sync()
	cerr := f.Close()
	if err := firstErr(werr, serr, cerr); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("evidence finalize %s: %w", tmp, err)
	}
	return nil
}

func firstErr(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}

// TempPathFor exposes the temp-file path the atomic writer uses so the
// repair subcommand can scrub orphans without re-deriving the suffix.
func TempPathFor(feature string, role Role) string {
	return filepath.Join(workflow.WorkflowDir, fmt.Sprintf("%s-%s.json%s", feature, role, tempSuffix))
}

// WriteBytesAtomic is the exported wrapper around writeBytesAtomic. Lets
// non-schema callers (postwrite hook, artifact templates from cmd/) reuse
// the temp-file + fsync + rename ritual without duplicating it.
func WriteBytesAtomic(target string, data []byte) error {
	return writeBytesAtomic(target, data)
}
