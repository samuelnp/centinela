package verify

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/evidence"
)

func writeFile(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCheckStubs(t *testing.T) {
	emptyTest := "package p\nimport \"testing\"\nfunc TestFoo(t *testing.T) {}\n"
	noAssert := "package p\nimport \"testing\"\nfunc TestBar(t *testing.T) {\n x := 1\n _ = x\n}\n"
	real := "package p\nimport \"testing\"\nfunc TestBaz(t *testing.T) {\n if 1 != 1 { t.Fatal(\"no\") }\n}\n"
	iface := "package verify\ntype CommandRunner interface { Run() }\n"
	blank := "package p\n// only a comment\n"

	cases := []struct {
		name string
		rel  string
		body string
		want Status
	}{
		{"empty-test-body-fail", "tests/unit/foo_test.go", emptyTest, StatusFail},
		{"zero-assertion-fail", "tests/unit/bar_test.go", noAssert, StatusFail},
		{"real-assertions-pass", "tests/unit/baz_test.go", real, StatusPass},
		{"tiny-interface-pass", "internal/verify/runner.go", iface, StatusPass},
		{"blank-nontest-fail", "internal/verify/blank.go", blank, StatusFail},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			writeFile(t, root, tc.rel, tc.body)
			ev := &evidence.RoleEvidence{Outputs: []string{tc.rel}}
			got := checkStubs(root, "qa", ev)
			if got.Status != tc.want {
				t.Fatalf("status = %q want %q (detail %q)", got.Status, tc.want, got.Detail)
			}
			if tc.want == StatusFail && !strings.Contains(got.Detail, tc.rel) {
				t.Fatalf("fail detail should name file, got %q", got.Detail)
			}
		})
	}
}

func TestCheckStubsEmptyAndNonGo(t *testing.T) {
	if got := checkStubs(t.TempDir(), "qa", &evidence.RoleEvidence{}); got.Status != StatusSkip {
		t.Fatalf("no outputs should skip, got %q", got.Status)
	}
	// Non-Go and unreadable files are conservatively treated as not-a-stub.
	ev := &evidence.RoleEvidence{Outputs: []string{"docs/x.md", "internal/verify/gone.go"}}
	if got := checkStubs(t.TempDir(), "qa", ev); got.Status != StatusPass {
		t.Fatalf("non-go + missing should pass, got %q / %q", got.Status, got.Detail)
	}
}
