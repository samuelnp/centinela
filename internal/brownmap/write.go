package brownmap

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// WriteDraft writes the plan's draft Roadmap to path as indented JSON, byte-for-
// byte identical to roadmap.Save's format (MarshalIndent two-space). It REFUSES
// to write the canonical roadmap.RoadmapFile so a curated roadmap is never
// clobbered, and writes atomically (temp file in the same dir + rename) so a
// failure leaves no partial draft and concurrent readers never see a half file.
// It returns the path actually written.
func WriteDraft(path string, p Plan) (wrote string, err error) {
	if filepath.Clean(path) == filepath.Clean(roadmap.RoadmapFile) {
		return "", fmt.Errorf("refusing to write brownfield draft to canonical roadmap %q", roadmap.RoadmapFile)
	}
	data, err := json.MarshalIndent(p.Roadmap, "", "  ")
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err = os.MkdirAll(dir, 0o755); err != nil {
			return "", err
		}
	}
	if err = atomicWrite(dir, path, data); err != nil {
		return "", err
	}
	return path, nil
}

// atomicWrite writes data to a temp file in dir, then renames it over path so the
// replacement is atomic on the same filesystem. The temp file is cleaned up on
// any failure before the rename.
func atomicWrite(dir, path string, data []byte) error {
	if dir == "" {
		dir = "."
	}
	tmp, err := os.CreateTemp(dir, ".roadmap.brownfield-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err = tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err = tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err = os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}
