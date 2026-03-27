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
	return nil
}
