package memory

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

// TestRegenerateIndexFromEntries — index.json reflects written entry files.
func TestRegenerateIndexFromEntries(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	e := newEntry("alpha", "tests", TypeLesson, "- body", "src", []string{"lesson"}, time.Now())
	if _, err := writeIfAbsent(e); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if err := regenerateIndex(); err != nil {
		t.Fatalf("regenerateIndex failed: %v", err)
	}
	data, err := os.ReadFile(indexFile)
	if err != nil {
		t.Fatalf("index file missing: %v", err)
	}
	var records []indexRecord
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].ID != e.ID {
		t.Fatalf("record ID mismatch: %q != %q", records[0].ID, e.ID)
	}
}

// TestRegenerateIndexEmptyLedger — writes empty JSON array for empty ledger.
func TestRegenerateIndexEmptyLedger(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	if err := regenerateIndex(); err != nil {
		t.Fatalf("regenerateIndex failed on empty ledger: %v", err)
	}
	data, err := os.ReadFile(indexFile)
	if err != nil {
		t.Fatalf("index file missing: %v", err)
	}
	var records []indexRecord
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("expected empty records, got %d", len(records))
	}
}

// TestRegenerateIndexSortedByID — records are sorted by ID.
func TestRegenerateIndexSortedByID(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	e1 := newEntry("b", "tests", TypeLesson, "- body b", "src", []string{}, time.Now())
	e2 := newEntry("a", "tests", TypeLesson, "- body a", "src", []string{}, time.Now())
	writeIfAbsent(e1) //nolint:errcheck
	writeIfAbsent(e2) //nolint:errcheck
	regenerateIndex() //nolint:errcheck

	data, _ := os.ReadFile(indexFile)
	var records []indexRecord
	json.Unmarshal(data, &records) //nolint:errcheck
	if len(records) == 2 && records[0].ID > records[1].ID {
		t.Fatalf("expected records sorted by ID, got %q > %q", records[0].ID, records[1].ID)
	}
}
