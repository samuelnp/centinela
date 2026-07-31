package docstring

import "testing"

func TestInScope_AcceptsSourceUnderRootsAndRejectsTests(t *testing.T) {
	opts := Options{IncludeInternal: true}
	cases := map[string]bool{
		"internal/config/config.go":  true,
		"cmd/centinela/main.go":      true,
		"internal/config/a_test.go":  false,
		"internal/config/config.txt": false,
		"web/src/app.go":             false,
		"README.md":                  false,
	}
	for path, want := range cases {
		if got := InScope(path, opts); got != want {
			t.Errorf("InScope(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestInScope_IncludeInternalGatesInternalPackages(t *testing.T) {
	p := "internal/config/config.go"
	if InScope(p, Options{IncludeInternal: false}) {
		t.Fatal("include_internal=false must exclude internal packages")
	}
	if !InScope(p, Options{IncludeInternal: true}) {
		t.Fatal("include_internal=true must include internal packages")
	}
	if !InScope("pkg/lib/x.go", Options{IncludeInternal: false}) {
		t.Fatal("non-internal path must stay in scope")
	}
}

func TestInScope_CustomRootsAndWholeTreeRoot(t *testing.T) {
	if !InScope("api/v1/x.go", Options{Roots: []string{"api"}, IncludeInternal: true}) {
		t.Fatal("custom root must be honored")
	}
	if InScope("cmd/x.go", Options{Roots: []string{"api"}, IncludeInternal: true}) {
		t.Fatal("path outside the custom roots must be excluded")
	}
	if !InScope("anywhere/x.go", Options{Roots: []string{"."}, IncludeInternal: true}) {
		t.Fatal(`root "." must accept the whole tree`)
	}
	if !InScope("cmd/x.go", Options{Roots: []string{" ", "cmd/"}, IncludeInternal: true}) {
		t.Fatal("roots must be trimmed of blanks and trailing slashes")
	}
}

func TestSelected_DeduplicatesAndDropsOutOfScopePaths(t *testing.T) {
	got := selected([]string{"cmd/a.go", "cmd/a.go", "README.md", "cmd/b_test.go"},
		Options{IncludeInternal: true})
	if len(got) != 1 || got[0] != "cmd/a.go" {
		t.Fatalf("selected = %v", got)
	}
}
