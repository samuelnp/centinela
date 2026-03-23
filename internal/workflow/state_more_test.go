package workflow

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveWriteErrorWhenTargetIsDir(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll(WorkflowDir, 0755)   //nolint:errcheck
	os.MkdirAll(FilePath("f"), 0755) //nolint:errcheck
	if err := Save(New("f")); err == nil {
		t.Fatal("expected write error when workflow file path is directory")
	}

	os.RemoveAll(FilePath("f"))                                            //nolint:errcheck
	os.WriteFile(filepath.Join(WorkflowDir, "x.json"), []byte("{}"), 0644) //nolint:errcheck
}
