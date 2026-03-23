package workflow

import (
	"os"
	"testing"
)

func TestStateNewSaveLoadFilePath(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	wf := New("feat")
	if wf.CurrentStep != "plan" || wf.Steps["plan"].Status != "in-progress" {
		t.Fatalf("unexpected new workflow: %+v", wf)
	}
	os.MkdirAll(WorkflowDir, 0755) //nolint:errcheck
	if err := Save(wf); err != nil {
		t.Fatalf("Save error: %v", err)
	}
	if FilePath("feat") != ".workflow/feat.json" {
		t.Fatalf("bad filepath: %s", FilePath("feat"))
	}
	got, err := Load("feat")
	if err != nil || got.Feature != "feat" {
		t.Fatalf("Load error: %v %+v", err, got)
	}
}

func TestLoadMissingWorkflow(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	if _, err := Load("missing"); err == nil {
		t.Fatal("expected missing workflow error")
	}
}

func TestSaveErrorWhenWorkflowDirConflicts(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                            //nolint:errcheck
	os.Chdir(d)                                  //nolint:errcheck
	os.WriteFile(WorkflowDir, []byte("x"), 0644) //nolint:errcheck
	if err := Save(New("f")); err == nil {
		t.Fatal("expected save error with conflicting workflow dir file")
	}
}

func TestLoadParseError(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                   //nolint:errcheck
	os.Chdir(d)                                         //nolint:errcheck
	os.MkdirAll(WorkflowDir, 0755)                      //nolint:errcheck
	os.WriteFile(FilePath("bad"), []byte("{bad"), 0644) //nolint:errcheck
	if _, err := Load("bad"); err == nil {
		t.Fatal("expected parse error for invalid workflow json")
	}
}
