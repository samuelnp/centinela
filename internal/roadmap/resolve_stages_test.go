package roadmap

import (
	"errors"
	"strings"
	"testing"
)

// stageStub answers git ls-files/show from a fixed table.
func stageStub(unmerged string, blobs map[string]string) StageRunner {
	return func(_ string, args ...string) (string, error) {
		switch args[0] {
		case "ls-files":
			return unmerged, nil
		case "show":
			if body, ok := blobs[args[1]]; ok {
				return body, nil
			}
			return "", errors.New("fatal: path does not exist")
		}
		return "", nil
	}
}

func TestReadStagesReadsAllThree(t *testing.T) {
	run := stageStub("100644 abc 1\t.workflow/roadmap.json\n", map[string]string{
		":1:.workflow/roadmap.json": "base",
		":2:.workflow/roadmap.json": "ours",
		":3:.workflow/roadmap.json": "theirs",
	})
	st, ok, err := ReadStages(".", ".workflow/roadmap.json", run)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if string(st.Base) != "base" || string(st.Ours) != "ours" || string(st.Theirs) != "theirs" {
		t.Fatalf("stages = %+v", st)
	}
}

// An absent stage 1 (both sides added the file) is normal, not an error.
func TestReadStagesToleratesAnAbsentBase(t *testing.T) {
	run := stageStub("100644 abc 2\t.workflow/roadmap.json\n", map[string]string{
		":2:.workflow/roadmap.json": "ours",
		":3:.workflow/roadmap.json": "theirs",
	})
	st, ok, err := ReadStages(".", ".workflow/roadmap.json", run)
	if err != nil || !ok || len(st.Base) != 0 {
		t.Fatalf("st=%+v ok=%v err=%v", st, ok, err)
	}
}

// E13: no unmerged entry means there is nothing to resolve.
func TestReadStagesReportsNoConflict(t *testing.T) {
	run := stageStub("", nil)
	if Conflicted(".", ".workflow/roadmap.json", run) {
		t.Fatal("an empty ls-files means no conflict")
	}
	_, ok, err := ReadStages(".", ".workflow/roadmap.json", run)
	if ok || err != nil {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}

func TestConflictedIsFalseWhenGitFails(t *testing.T) {
	run := StageRunner(func(string, ...string) (string, error) { return "", errors.New("not a repo") })
	if Conflicted(".", ".workflow/roadmap.json", run) {
		t.Fatal("a git failure must not be read as a conflict")
	}
}

// Conflicted with no readable content on either side is a real error, not a
// silent merge of two empty documents.
func TestReadStagesRefusesAContentlessConflict(t *testing.T) {
	run := stageStub("100644 abc 1\t.workflow/roadmap.json\n", map[string]string{
		":1:.workflow/roadmap.json": "base",
	})
	_, _, err := ReadStages(".", ".workflow/roadmap.json", run)
	if err == nil || !strings.Contains(err.Error(), "neither side") {
		t.Fatalf("want a contentless-conflict error, got %v", err)
	}
}
