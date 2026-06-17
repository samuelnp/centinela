package doctor

import (
	"strings"
	"testing"
)

func seedMakefile(t *testing.T, dir, ver string) {
	t.Helper()
	writeFile(t, "Makefile", "VERSION := "+ver+"\n\nbuild:\n\tgo build\n")
}

func TestVersionCheckMatchOK(t *testing.T) {
	dir := repoFixture(t)
	seedMakefile(t, dir, "0.21.1")
	stubVersion(t, func() (string, error) { return "centinela version 0.21.1\n", nil })
	d := versionCheck{}.Run(Context{Root: dir})
	if d.Status != OK {
		t.Fatalf("matching version must be OK, got %v %q", d.Status, d.Message)
	}
}

func TestVersionCheckBehindWarn(t *testing.T) {
	dir := repoFixture(t)
	seedMakefile(t, dir, "0.21.1")
	stubVersion(t, func() (string, error) { return "centinela version 0.15.0\n", nil })
	d := versionCheck{}.Run(Context{Root: dir})
	if d.Status != Warn {
		t.Fatalf("behind must Warn, got %v", d.Status)
	}
	if !strings.Contains(d.Message, "0.15.0") || !strings.Contains(d.Message, "0.21.1") {
		t.Fatalf("message must report both versions: %q", d.Message)
	}
	if d.Repair == nil || d.Repair.Command != "make install" {
		t.Fatalf("repair must recommend make install, got %v", d.Repair)
	}
}

func TestVersionCheckBinaryNotFoundWarn(t *testing.T) {
	dir := repoFixture(t)
	seedMakefile(t, dir, "0.21.1")
	stubVersion(t, func() (string, error) { return "", errStub })
	d := versionCheck{}.Run(Context{Root: dir})
	if d.Status != Warn || !strings.Contains(d.Message, "not found") {
		t.Fatalf("not found must Warn, got %v %q", d.Status, d.Message)
	}
	if d.Repair != nil {
		t.Fatal("not-found degrade carries no repair")
	}
}

func TestMakefileVersion(t *testing.T) {
	dir := repoFixture(t)
	seedMakefile(t, dir, "1.2.3")
	if got := makefileVersion(dir); got != "1.2.3" {
		t.Fatalf("makefileVersion=%q", got)
	}
	if got := makefileVersion(t.TempDir()); got != "" {
		t.Fatalf("missing Makefile must be empty, got %q", got)
	}
}
