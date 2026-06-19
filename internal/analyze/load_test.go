package analyze

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "analysis.json")
	want := Inventory{SchemaVersion: SchemaVersion, PrimaryLanguage: "Go", Packages: []string{"a", "b"}}
	if err := Save(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil || got.PrimaryLanguage != "Go" || len(got.Packages) != 2 {
		t.Fatalf("round-trip mismatch: %+v err=%v", got, err)
	}
}

func TestLoad_MissingFileIsErrNoInventory(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "absent.json"))
	if !errors.Is(err, ErrNoInventory) {
		t.Fatalf("missing file must be ErrNoInventory, got %v", err)
	}
}

func TestLoad_MalformedJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	os.WriteFile(path, []byte("{not json"), 0o644)
	if _, err := Load(path); err == nil || errors.Is(err, ErrNoInventory) {
		t.Fatalf("malformed JSON must be a distinct error, got %v", err)
	}
}

func TestLoad_ReadErrorNotMissing(t *testing.T) {
	// A directory path makes os.ReadFile fail with a non-IsNotExist error, so
	// Load returns the "reading" error rather than ErrNoInventory.
	_, err := Load(t.TempDir())
	if err == nil || errors.Is(err, ErrNoInventory) {
		t.Fatalf("reading a directory must be a distinct read error, got %v", err)
	}
}

func TestLoad_SchemaDrift(t *testing.T) {
	path := filepath.Join(t.TempDir(), "old.json")
	os.WriteFile(path, []byte(`{"schemaVersion": 999}`), 0o644)
	_, err := Load(path)
	if err == nil || errors.Is(err, ErrNoInventory) {
		t.Fatalf("schema drift must be a distinct error, got %v", err)
	}
}
