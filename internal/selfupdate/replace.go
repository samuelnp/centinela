package selfupdate

import (
	"os"
	"path/filepath"
)

// Overridable seams (mirrors the gitutil leaf pattern) so the resolver's error
// branches are testable without a real broken install.
var (
	osExecutable = os.Executable
	evalSymlinks = filepath.EvalSymlinks
	writeTempFn  = writeTemp
	syncFn       = (*os.File).Sync
	chmodFn      = (*os.File).Chmod
)

// targetPath resolves the real path of the running binary, following symlinks so
// a symlinked install is replaced at its true location.
func targetPath() (string, error) {
	exe, err := osExecutable()
	if err != nil {
		return "", newErr(KindReplace, "resolve executable", err)
	}
	real, err := evalSymlinks(exe)
	if err != nil {
		return "", newErr(KindReplace, "resolve symlink", err)
	}
	return real, nil
}

// replaceBinary atomically swaps the file at target with data: it writes a temp
// file in the SAME directory, fsyncs, copies the existing mode bits, then renames
// over target. On any failure the temp file is removed and target is untouched.
func replaceBinary(target string, data []byte) error {
	info, err := os.Stat(target)
	if err != nil {
		return newErr(KindReplace, "stat target binary", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(target), ".centinela-update-*")
	if err != nil {
		return newErr(KindPermission, "create temp file in install directory", err)
	}
	name := tmp.Name()
	if err := writeTempFn(tmp, data, info.Mode()); err != nil {
		os.Remove(name)
		return err
	}
	if err := os.Rename(name, target); err != nil {
		os.Remove(name)
		return newErr(KindReplace, "rename temp over target", err)
	}
	return nil
}

// writeTemp writes data to tmp, fsyncs, applies mode, and closes it.
func writeTemp(tmp *os.File, data []byte, mode os.FileMode) error {
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return newErr(KindReplace, "write temp file", err)
	}
	if err := syncFn(tmp); err != nil {
		tmp.Close()
		return newErr(KindReplace, "fsync temp file", err)
	}
	if err := chmodFn(tmp, mode); err != nil {
		tmp.Close()
		return newErr(KindReplace, "chmod temp file", err)
	}
	return tmp.Close()
}
