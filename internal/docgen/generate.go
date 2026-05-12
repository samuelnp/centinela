package docgen

import (
	"fmt"
	"os"
	"path/filepath"
)

func Generate(outPath, title string) error {
	d, err := LoadData(title)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return fmt.Errorf("cannot create output dir: %w", err)
	}
	if err := os.WriteFile(outPath, []byte(RenderHTML(d)), 0644); err != nil {
		return fmt.Errorf("cannot write html: %w", err)
	}
	return writeKB(d, title)
}

func writeKB(d *Data, title string) error {
	if err := os.MkdirAll(KBDir, 0755); err != nil {
		return fmt.Errorf("cannot create kb dir: %w", err)
	}
	idx := filepath.Join(KBDir, "index.html")
	if err := os.WriteFile(idx, []byte(RenderKBIndex(d)), 0644); err != nil {
		return fmt.Errorf("cannot write kb index: %w", err)
	}
	for _, p := range d.KB {
		path := filepath.Join(KBDir, p.Feature+".html")
		if err := os.WriteFile(path, []byte(RenderKBFeature(p, title)), 0644); err != nil {
			return fmt.Errorf("cannot write kb page %s: %w", p.Feature, err)
		}
	}
	return nil
}
