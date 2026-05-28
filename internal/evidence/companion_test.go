package evidence

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func TestWriteCompanionAndRead(t *testing.T) {
	chdirToTemp(t)
	body := DefaultCompanionTemplate("alpha", orchestration.RoleBigThinker)
	if err := WriteCompanion("alpha", orchestration.RoleBigThinker, body); err != nil {
		t.Fatal(err)
	}
	got, err := ReadCompanion("alpha", orchestration.RoleBigThinker)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "alpha — big-thinker") {
		t.Fatalf("unexpected companion: %q", got)
	}
	path := companionPath("alpha", orchestration.RoleBigThinker)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("companion not on disk: %v", err)
	}
}

func TestReadCompanionMissingIsEmpty(t *testing.T) {
	chdirToTemp(t)
	got, err := ReadCompanion("ghost", orchestration.RoleBigThinker)
	if err != nil || got != "" {
		t.Fatalf("expected empty no-error, got %q,%v", got, err)
	}
}

func TestDefaultCompanionTemplateMentionsRole(t *testing.T) {
	body := DefaultCompanionTemplate("alpha", orchestration.RoleFeatureSpecial)
	if !strings.Contains(body, "feature-specialist") {
		t.Fatalf("template missing role: %q", body)
	}
}
