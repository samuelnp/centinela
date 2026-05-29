package memory

import (
	"os"
	"path/filepath"
)

const (
	ledgerDir  = ".workflow/memory"
	entriesDir = ".workflow/memory/entries"
	indexFile  = ".workflow/memory/index.json"
)

// entryPath returns the on-disk path for an entry id.
func entryPath(id string) string {
	return filepath.Join(entriesDir, id+".md")
}

// writeIfAbsent writes the entry to its own file only when it does not yet
// exist, making capture idempotent (SC-05) and concurrency-safe (SC-13).
// It reports whether a new file was written.
func writeIfAbsent(e Entry) (bool, error) {
	if err := os.MkdirAll(entriesDir, 0o755); err != nil {
		return false, err
	}
	path := entryPath(e.ID)
	if _, err := os.Stat(path); err == nil {
		return false, nil
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return false, nil
		}
		return false, err
	}
	defer f.Close()
	if _, err := f.Write(marshal(e)); err != nil {
		return false, err
	}
	return true, nil
}

// loadEntries reads every entry file from the ledger. Malformed files are
// skipped so a single bad file never breaks recall.
func loadEntries() []Entry {
	dir, err := os.ReadDir(entriesDir)
	if err != nil {
		return nil
	}
	out := []Entry{}
	for _, d := range dir {
		if d.IsDir() || filepath.Ext(d.Name()) != ".md" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(entriesDir, d.Name()))
		if err != nil {
			continue
		}
		if e, ok := unmarshal(data); ok {
			out = append(out, e)
		}
	}
	return out
}
