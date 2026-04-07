package gates

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func TestCheckFileSizePassAndFail(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("src", 0755)                              //nolint:errcheck
	os.WriteFile("src/a.go", []byte("package a\n"), 0644) //nolint:errcheck
	if r := checkFileSize(&config.Config{}); r.Status != Pass {
		t.Fatalf("expected pass, got %v", r.Status)
	}
	big := ""
	for i := 0; i < 101; i++ {
		big += "x\n"
	}
	os.WriteFile("src/b.go", []byte(big), 0644) //nolint:errcheck
	r := checkFileSize(&config.Config{})
	if r.Status != Fail || len(r.Details) == 0 {
		t.Fatalf("expected fail with violations, got %+v", r)
	}
}

func TestExistingRootsAndFindOversized(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("cmd", 0755)                                                   //nolint:errcheck
	os.WriteFile(filepath.Join("cmd", "x.go"), []byte("package main\n"), 0644) //nolint:errcheck
	if roots := existingRoots(); len(roots) == 0 {
		t.Fatal("expected at least one existing root")
	}
	if v, _ := findOversizedFiles(&config.Config{}); len(v) != 0 {
		t.Fatalf("did not expect violations, got %v", v)
	}
}
