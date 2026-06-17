package gates

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// goListPkg mirrors the subset of `go list -json` output this gate consumes.
// Imports/TestImports/XTestImports together capture production code plus
// in-package and external (_test) test imports so test files cannot smuggle a
// forbidden cross-layer edge past the gate.
type goListPkg struct {
	ImportPath   string
	Imports      []string
	TestImports  []string
	XTestImports []string
}

// loadModulePath returns the current module path via `go list -m`.
func loadModulePath() (string, error) {
	out, err := runGo("list", "-m")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// loadPackages runs `go list -json ./...` and decodes the streamed (concatenated,
// not array) JSON objects, returning one goListPkg per package. A non-zero exit
// (e.g. uncompilable code) is surfaced as an error so the gate Fails rather than
// reporting a false Pass.
func loadPackages() ([]goListPkg, error) {
	out, err := runGo("list", "-json", "./...")
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(out))
	var pkgs []goListPkg
	for {
		var p goListPkg
		if derr := dec.Decode(&p); derr == io.EOF {
			break
		} else if derr != nil {
			return nil, fmt.Errorf("decoding go list output: %w", derr)
		}
		pkgs = append(pkgs, p)
	}
	return pkgs, nil
}

// runGo executes a go subcommand, returning stdout or an error whose message
// includes the first stderr line for actionable load diagnostics.
func runGo(args ...string) ([]byte, error) {
	cmd := exec.Command("go", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if line := firstStderrLine(stderr.String(), err); line != "" {
			return nil, fmt.Errorf("go %s: %s", strings.Join(args, " "), line)
		}
		return nil, fmt.Errorf("go %s: %w", strings.Join(args, " "), err)
	}
	return stdout.Bytes(), nil
}
