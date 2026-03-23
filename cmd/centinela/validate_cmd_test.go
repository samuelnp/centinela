package main

import (
	"os"
	"testing"
)

func TestExecuteValidationPassAndFail(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.WriteFile("centinela.toml", []byte("[gates]\nfile_size=true\n[validate]\ncommands=[\"true\"]\n"), 0644) //nolint:errcheck
	if err := executeValidation(); err != nil {
		t.Fatalf("executeValidation should pass: %v", err)
	}
	os.WriteFile("centinela.toml", []byte("[gates]\nfile_size=true\n[validate]\ncommands=[\"false\"]\n"), 0644) //nolint:errcheck
	if err := executeValidation(); err == nil {
		t.Fatal("expected validation failure")
	}
}
