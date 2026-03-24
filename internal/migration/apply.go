package migration

import (
	"os"
	"path/filepath"
)

func Apply(root string, plan Plan) error {
	for _, it := range plan.Items {
		target := filepath.Join(root, it.Path)
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(target, []byte(it.content), 0644); err != nil {
			return err
		}
	}
	return nil
}
