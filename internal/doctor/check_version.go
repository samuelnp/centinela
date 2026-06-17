package doctor

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// versionRunner is overridable for tests; default runs `centinela --version`.
var versionRunner = func() (string, error) {
	out, err := exec.Command("centinela", "--version").CombinedOutput()
	return string(out), err
}

var versionRe = regexp.MustCompile(`(\d+\.\d+\.\d+)`)

// versionCheck compares the installed centinela binary's version against the
// repo Makefile VERSION. WARN, report-only (doctor cannot safely self-install);
// the remediation is `make install`. Binary not found degrades to WARN, never a
// crash.
type versionCheck struct{}

func (versionCheck) Name() string { return "version" }

func (versionCheck) Run(ctx Context) Diagnosis {
	d := Diagnosis{Name: "version"}
	makeVer := makefileVersion(ctx.Root)
	out, err := versionRunner()
	if err != nil {
		d.Status = Warn
		d.Message = "centinela binary not found on PATH — run `make install`"
		return d
	}
	installed := versionRe.FindString(out)
	if installed == "" || makeVer == "" || installed == makeVer {
		d.Status = OK
		d.Message = "installed binary matches Makefile VERSION (" + makeVer + ")"
		return d
	}
	d.Status = Warn
	d.Message = "installed centinela " + installed + " is behind Makefile VERSION " + makeVer
	d.Details = []string{"run `make install` to update the installed binary"}
	d.Repair = &Repair{Command: "make install"}
	return d
}

// makefileVersion parses `VERSION := X.Y.Z` from the repo Makefile.
func makefileVersion(root string) string {
	data, err := os.ReadFile(filepath.Join(root, "Makefile"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "VERSION") {
			return versionRe.FindString(line)
		}
	}
	return ""
}
