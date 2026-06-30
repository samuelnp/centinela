package evidence

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func TestWriteBytesAtomicExportedHappyPath(t *testing.T) {
	d := chdirToTemp(t)
	target := filepath.Join(d, ".workflow", "exported.json")
	if err := WriteBytesAtomic(target, []byte(`{"ok":true}`)); err != nil {
		t.Fatalf("WriteBytesAtomic: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(got) != `{"ok":true}` {
		t.Fatalf("content mismatch: %q", got)
	}
	// The temp sibling must be gone after a successful rename.
	if _, err := os.Stat(target + tempSuffix); !os.IsNotExist(err) {
		t.Fatalf("temp file not cleaned up: %v", err)
	}
}

func TestWriteBytesAtomicExportedErrorWhenTempPathIsDir(t *testing.T) {
	d := chdirToTemp(t)
	target := filepath.Join(d, ".workflow", "blocked.json")
	// A directory squatting on the .tmp sibling makes the temp open fail.
	if err := os.MkdirAll(target+tempSuffix, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := WriteBytesAtomic(target, []byte("x")); err == nil {
		t.Fatal("expected error when temp path is a directory")
	}
}

func TestSuggestCommandIncompleteField(t *testing.T) {
	got := suggestCommand("alpha", orchestration.RoleBigThinker, FieldError{Field: "incomplete"})
	if got == "" {
		t.Fatal("incomplete field should yield a fix command")
	}
}

func TestMarshalNoEscapeEncodeError(t *testing.T) {
	// Channels are not JSON-encodable, exercising the error return.
	if _, err := marshalNoEscape(make(chan int)); err == nil {
		t.Fatal("expected encode error for chan value")
	}
}
