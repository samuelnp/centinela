package workflow

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func TestValidateTestsRejectsPlaceholderOnly(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                             //nolint:errcheck
	os.Chdir(d)                                                   //nolint:errcheck
	os.MkdirAll("tests/unit", 0755)                               //nolint:errcheck
	os.MkdirAll("tests/acceptance", 0755)                         //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                //nolint:errcheck
	os.WriteFile("tests/unit/.gitkeep", []byte("x"), 0644)        //nolint:errcheck
	os.WriteFile("tests/acceptance/.gitkeep", []byte("x"), 0644)  //nolint:errcheck
	os.WriteFile(".workflow/f-edge-cases.md", []byte("ok"), 0644) //nolint:errcheck
	if err := validateTests("f", &config.Config{}); err == nil {
		t.Fatal("expected failure with placeholder-only test artifacts")
	}
}
