// Package golist is a tiny leaf wrapper around the `go list` toolchain that
// loads a module's package import graph and module path. It is the single
// shared seam reused by both the import_graph gate and codebase analysis so the
// streamed-JSON decode and stderr-surfacing logic is not duplicated. It depends
// only on the standard library + os/exec.
package golist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Pkg mirrors the subset of `go list -json` output callers consume.
// Imports/TestImports/XTestImports together capture production code plus
// in-package and external (_test) test imports so test files cannot hide a
// forbidden cross-layer edge.
type Pkg struct {
	ImportPath   string
	Imports      []string
	TestImports  []string
	XTestImports []string
}

// ModulePath returns the current module path via `go list -m`.
func ModulePath() (string, error) {
	out, err := runGo("list", "-m")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// Packages runs `go list -json ./...` and decodes the streamed (concatenated,
// not array) JSON objects, returning one Pkg per package. A non-zero exit
// (e.g. uncompilable code) is surfaced as an error so callers never treat
// unloadable code as an empty-but-valid graph.
func Packages() ([]Pkg, error) {
	out, err := runGo("list", "-json", "./...")
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(out))
	var pkgs []Pkg
	for {
		var p Pkg
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
		if line := firstStderrLine(stderr.String()); line != "" {
			return nil, fmt.Errorf("go %s: %s", strings.Join(args, " "), line)
		}
		return nil, fmt.Errorf("go %s: %w", strings.Join(args, " "), err)
	}
	return stdout.Bytes(), nil
}

// firstStderrLine returns the first non-empty line of stderr, used to fold a
// concise toolchain diagnostic into the returned error.
func firstStderrLine(stderr string) string {
	for _, line := range strings.Split(stderr, "\n") {
		if s := strings.TrimSpace(line); s != "" {
			return s
		}
	}
	return ""
}
