package synthesize

import (
	"os"
	"path/filepath"
)

// DefaultTarget is the canonical project-definition path. It is never
// overwritten; a draft is written alongside it instead.
const DefaultTarget = "PROJECT.md"

// DraftTarget is where the synthesized draft is written when DefaultTarget
// already exists.
const DraftTarget = "PROJECT.draft.md"

// WriteDraft writes content to target, but never overwrites an existing
// DefaultTarget: when target already exists it writes to DraftTarget instead and
// reports clobbered=true. Returns the path actually written. The payload is
// written in one call so a failure leaves no partial file.
func WriteDraft(target, content string) (written string, clobbered bool, err error) {
	out := target
	if _, statErr := os.Stat(target); statErr == nil {
		out = draftPathFor(target)
		clobbered = true
	}
	if dir := filepath.Dir(out); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", clobbered, err
		}
	}
	if err := os.WriteFile(out, []byte(content), 0o644); err != nil {
		return "", clobbered, err
	}
	return out, clobbered, nil
}

// draftPathFor returns the sibling draft path for an existing target, e.g.
// PROJECT.md -> PROJECT.draft.md, dir/PROJECT.md -> dir/PROJECT.draft.md.
func draftPathFor(target string) string {
	dir, base := filepath.Split(target)
	ext := filepath.Ext(base)
	stem := base[:len(base)-len(ext)]
	return filepath.Join(dir, stem+".draft"+ext)
}
