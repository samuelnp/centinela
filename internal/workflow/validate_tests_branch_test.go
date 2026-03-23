package workflow

import (
	"os"
	"testing"
)

func TestHasUnitOrIntegrationTestsSuffixBranches(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("tests/unit", 0755)                         //nolint:errcheck
	os.WriteFile("tests/unit/a_test.go", []byte("x"), 0644) //nolint:errcheck
	if !hasUnitOrIntegrationTests([]string{"_test.go"}) {
		t.Fatal("expected suffix match in unit tests")
	}
	os.Remove("tests/unit/a_test.go")                              //nolint:errcheck
	os.MkdirAll("tests/integration", 0755)                         //nolint:errcheck
	os.WriteFile("tests/integration/a_spec.rb", []byte("x"), 0644) //nolint:errcheck
	if !hasUnitOrIntegrationTests([]string{"_spec.rb"}) {
		t.Fatal("expected suffix match in integration tests")
	}
}
