package main

import (
	"os"
	"testing"
)

func TestRunValidateWrapper(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.WriteFile("centinela.toml", []byte("[gates]\nfile_size=true\n"), 0644) //nolint:errcheck
	if err := runValidate(nil, nil); err != nil {
		t.Fatalf("runValidate should pass: %v", err)
	}
}

func TestRunCommandWindowsBranch(t *testing.T) {
	old := runtimeOS
	defer func() { runtimeOS = old }()
	runtimeOS = "windows"
	ok, _ := runCommand("echo hi")
	if ok {
		t.Fatal("expected windows branch command to fail on non-windows host")
	}
}
