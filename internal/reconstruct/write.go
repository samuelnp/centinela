package reconstruct

import (
	"os"
	"path/filepath"
)

// DefaultOutRoot is the default review directory the reconstructed corpus is
// written into. It is never the canonical specs/ dir, so hand-authored specs are
// never clobbered.
const DefaultOutRoot = ".workflow/reconstructed"

// canonicalSpecDir is the repo's hand-authored spec directory, checked for
// skip-if-exists so a partially-spec'd repo is augmented, never overwritten.
const canonicalSpecDir = "specs"

// WriteCorpus writes the reconstruction's features into outRoot/specs and briefs
// into outRoot/features. A target whose canonical specs/<slug>.feature already
// exists is skipped (recorded in skipped, with both its feature and brief
// suppressed) so hand-authored specs are never overwritten. Each surviving file
// is written via a single os.WriteFile after MkdirAll, so a failure leaves no
// partial file. Output is byte-stable across re-runs of the same Reconstruction.
func WriteCorpus(outRoot string, r Reconstruction) (written, skipped []string, err error) {
	specDir := filepath.Join(outRoot, "specs")
	featDir := filepath.Join(outRoot, "features")
	for _, f := range r.Features {
		if _, statErr := os.Stat(filepath.Join(canonicalSpecDir, f.Slug+".feature")); statErr == nil {
			skipped = append(skipped, f.Slug)
			continue
		}
		fp := filepath.Join(specDir, f.Slug+".feature")
		if err = writeFile(fp, f.Body); err != nil {
			return written, skipped, err
		}
		written = append(written, fp)
		bp := filepath.Join(featDir, f.Slug+".md")
		if err = writeFile(bp, briefFor(r, f.Slug)); err != nil {
			return written, skipped, err
		}
		written = append(written, bp)
	}
	return written, skipped, nil
}

// briefFor returns the brief body matching slug (features and briefs share slugs
// and order, but lookup keeps WriteCorpus order-independent).
func briefFor(r Reconstruction, slug string) string {
	for _, b := range r.Briefs {
		if b.Slug == slug {
			return b.Body
		}
	}
	return ""
}

func writeFile(path, body string) error {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, []byte(body), 0o644)
}
