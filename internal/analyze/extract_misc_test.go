package analyze

import "testing"

func TestExtractGoMod_ModulePath(t *testing.T) {
	var m Manifest
	extractGoMod([]byte("// c\nmodule example.com/foo\n\ngo 1.21\n"), &m)
	if m.Build != "example.com/foo" {
		t.Fatalf("module path: %q", m.Build)
	}
}

func TestExtractCargo_Deps(t *testing.T) {
	var m Manifest
	extractCargo([]byte("[package]\nname=\"x\"\n[dependencies]\nserde = \"1\"\ntokio = \"1\"\n"), &m)
	if len(m.Deps) != 2 || m.Deps[0] != "serde" || m.Deps[1] != "tokio" {
		t.Fatalf("cargo deps: %v", m.Deps)
	}
}

func TestExtractGemfile_Gems(t *testing.T) {
	var m Manifest
	extractGemfile([]byte("source \"x\"\ngem \"rails\"\ngem 'rspec'\n"), &m)
	if len(m.Deps) != 2 || m.Deps[0] != "rails" || m.Deps[1] != "rspec" {
		t.Fatalf("gemfile deps: %v", m.Deps)
	}
}

func TestExtractPyproject_PoetryDeps(t *testing.T) {
	var m Manifest
	src := "[tool.poetry.dependencies]\npython = \"^3.11\"\nrequests = \"^2\"\n"
	extractPyproject([]byte(src), &m)
	if len(m.Deps) != 1 || m.Deps[0] != "requests" {
		t.Fatalf("pyproject deps (python excluded): %v", m.Deps)
	}
}

func TestExtractRequirements_StripsSpecifiers(t *testing.T) {
	var m Manifest
	extractRequirements([]byte("# c\nflask==2.0\nrequests>=1\n\nnumpy\n"), &m)
	want := map[string]bool{"flask": true, "requests": true, "numpy": true}
	if len(m.Deps) != 3 {
		t.Fatalf("requirements deps: %v", m.Deps)
	}
	for _, d := range m.Deps {
		if !want[d] {
			t.Fatalf("unexpected dep %q in %v", d, m.Deps)
		}
	}
}

func TestExtractMakefile_BuildTestTargets(t *testing.T) {
	var m Manifest
	extractMakefile([]byte("all: build\nbuild:\n\tgo build\n"), &m)
	if m.Build != "make build" || m.Test != "" {
		t.Fatalf("makefile signals: %#v", m)
	}
}
