package analyze

import "strings"

// extractGoMod records the declared module path as the manifest framework-free
// build identity. The module path is the first `module <path>` directive.
func extractGoMod(data []byte, m *Manifest) {
	for _, line := range strings.Split(string(data), "\n") {
		s := strings.TrimSpace(line)
		if rest, ok := strings.CutPrefix(s, "module "); ok {
			m.Build = strings.TrimSpace(rest)
			return
		}
	}
}

// extractMakefile records build/test signals when the Makefile declares the
// conventional `build:` / `test:` targets.
func extractMakefile(data []byte, m *Manifest) {
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "build:") {
			m.Build = "make build"
		}
		if strings.HasPrefix(line, "test:") {
			m.Test = "make test"
		}
	}
}

// extractCargo records the package name as build identity and any declared
// dependency names from the [dependencies] table (best-effort line scan).
func extractCargo(data []byte, m *Manifest) {
	m.Deps = sortedDepNames(tomlSection(data, "dependencies"))
}

// extractGemfile records declared gem names from `gem "name"` lines.
func extractGemfile(data []byte, m *Manifest) {
	var names []string
	for _, line := range strings.Split(string(data), "\n") {
		s := strings.TrimSpace(line)
		if rest, ok := strings.CutPrefix(s, "gem "); ok {
			if n := firstQuoted(rest); n != "" {
				names = append(names, n)
			}
		}
	}
	m.Deps = sortedDepNames(asSet(names))
}

// extractPyproject records dependency names from a [project] dependencies array
// or a [tool.poetry.dependencies] table (best-effort).
func extractPyproject(data []byte, m *Manifest) {
	m.Deps = sortedDepNames(tomlSection(data, "tool.poetry.dependencies"))
}

// extractRequirements records one dependency per non-comment line, stripping
// any version specifier.
func extractRequirements(data []byte, m *Manifest) {
	var names []string
	for _, line := range strings.Split(string(data), "\n") {
		s := strings.TrimSpace(line)
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}
		names = append(names, splitReqName(s))
	}
	m.Deps = sortedDepNames(asSet(names))
}
