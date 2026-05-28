package evidence

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func TestReadCompanionReturnsBody(t *testing.T) {
	chdirToTemp(t)
	if err := WriteCompanion("alpha", orchestration.RoleBigThinker, "body"); err != nil {
		t.Fatal(err)
	}
	got, err := ReadCompanion("alpha", orchestration.RoleBigThinker)
	if err != nil || got != "body" {
		t.Fatalf("unexpected: %q,%v", got, err)
	}
}

func TestReadCompanionPropagatesUnexpectedErrors(t *testing.T) {
	d := chdirToTemp(t)
	// Create a directory at the companion path so ReadFile returns EISDIR.
	dir := filepath.Join(d, ".workflow", "alpha-big-thinker.md")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadCompanion("alpha", orchestration.RoleBigThinker); err == nil {
		t.Fatal("expected error reading directory as file")
	}
}
