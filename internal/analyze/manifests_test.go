package analyze

import (
	"path/filepath"
	"testing"
)

// findManifest returns the manifest of the given kind, or fails.
func findManifest(t *testing.T, ms []Manifest, kind string) Manifest {
	t.Helper()
	for _, m := range ms {
		if m.Kind == kind {
			return m
		}
	}
	t.Fatalf("manifest kind %q not detected in %#v", kind, ms)
	return Manifest{}
}

func TestDetectManifests_SortedByPath(t *testing.T) {
	root := t.TempDir()
	mkFile(t, filepath.Join(root, "go.mod"), "module example.com/m\n\ngo 1.21\n")
	mkFile(t, filepath.Join(root, "Makefile"), "build:\n\tgo build\ntest:\n\tgo test\n")
	mkFile(t, filepath.Join(root, "package.json"), `{"scripts":{"build":"vite build","test":"vitest"},"dependencies":{"react":"18"}}`)
	got := detectManifests(root)
	for i := 1; i < len(got); i++ {
		if got[i-1].Path > got[i].Path {
			t.Fatalf("manifests must be sorted by path: %#v", got)
		}
	}
	gomod := findManifest(t, got, "go-mod")
	if gomod.Path != "go.mod" || gomod.Build != "example.com/m" {
		t.Fatalf("go-mod module path: %#v", gomod)
	}
	mk := findManifest(t, got, "make")
	if mk.Build != "make build" || mk.Test != "make test" {
		t.Fatalf("make signals: %#v", mk)
	}
}

func TestDetectManifests_NpmScriptsAndDeps(t *testing.T) {
	root := t.TempDir()
	mkFile(t, filepath.Join(root, "package.json"),
		`{"scripts":{"build":"next build","test":"jest"},"dependencies":{"next":"14"},"devDependencies":{"jest":"29"}}`)
	npm := findManifest(t, detectManifests(root), "npm")
	if npm.Build != "next build" || npm.Test != "jest" {
		t.Fatalf("npm scripts: %#v", npm)
	}
	if npm.Framework != "Next.js" {
		t.Fatalf("npm framework: %q", npm.Framework)
	}
	if len(npm.Deps) != 2 || npm.Deps[0] != "jest" || npm.Deps[1] != "next" {
		t.Fatalf("npm deps must be sorted: %v", npm.Deps)
	}
}

func TestDetectManifests_MalformedPackageJSONStillDetected(t *testing.T) {
	root := t.TempDir()
	mkFile(t, filepath.Join(root, "package.json"), "{ this is not json")
	npm := findManifest(t, detectManifests(root), "npm")
	if npm.Build != "" || npm.Test != "" || len(npm.Deps) != 0 {
		t.Fatalf("malformed package.json must be detected-but-unparsable: %#v", npm)
	}
}

func TestDetectManifests_NoneWhenAbsent(t *testing.T) {
	if got := detectManifests(t.TempDir()); len(got) != 0 {
		t.Fatalf("no manifests expected, got %#v", got)
	}
}
