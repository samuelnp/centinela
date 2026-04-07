package workflow

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func TestValidateTestsRejectsCommentOnlyAcceptance(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                      //nolint:errcheck
	os.Chdir(d)                                                            //nolint:errcheck
	os.MkdirAll("tests/unit", 0755)                                        //nolint:errcheck
	os.MkdirAll("tests/acceptance", 0755)                                  //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                         //nolint:errcheck
	os.WriteFile("tests/unit/a_test.go", []byte("x"), 0644)                //nolint:errcheck
	os.WriteFile("tests/acceptance/a.go", []byte("// comment only"), 0644) //nolint:errcheck
	os.WriteFile(".workflow/f-edge-cases.md", []byte("ok"), 0644)          //nolint:errcheck
	err := validateTests("f", &config.Config{Validate: config.ValidateConfig{Commands: []string{"go test ./..."}}})
	if err == nil || !strings.Contains(err.Error(), "executable acceptance") {
		t.Fatalf("expected executable acceptance error, got %v", err)
	}
}

func TestValidateTestsRejectsMissingAcceptanceExecutionCommand(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                                                                                             //nolint:errcheck
	os.Chdir(d)                                                                                                                                   //nolint:errcheck
	os.MkdirAll("tests/unit", 0755)                                                                                                               //nolint:errcheck
	os.MkdirAll("tests/acceptance", 0755)                                                                                                         //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                                                                                                //nolint:errcheck
	os.WriteFile("tests/unit/a_test.go", []byte("x"), 0644)                                                                                       //nolint:errcheck
	os.WriteFile("tests/acceptance/a.go", []byte("package acceptance_test\nimport \"testing\"\nfunc TestA(t *testing.T){ t.Log(\"ok\") }"), 0644) //nolint:errcheck
	os.WriteFile(".workflow/f-edge-cases.md", []byte("ok"), 0644)                                                                                 //nolint:errcheck
	err := validateTests("f", &config.Config{Validate: config.ValidateConfig{Commands: []string{"go vet ./..."}}})
	if err == nil || !strings.Contains(err.Error(), "validate.commands") {
		t.Fatalf("expected acceptance execution command error, got %v", err)
	}
}
