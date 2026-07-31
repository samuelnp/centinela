package gates

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gitdiff"
)

func TestCheckFileSize_FilterEmptySetSkipsAllFiles(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("src", 0755) //nolint:errcheck
	big := makeBigSource(101)
	os.WriteFile("src/big.go", []byte(big), 0644) //nolint:errcheck

	r := checkFileSize(&config.Config{}, gitdiff.NewSet(nil))
	if r.Status != Skip {
		t.Fatalf("a gate that inspected nothing must report Skip, got %v: %v", r.Status, r.Details)
	}
	if r.Message != "No files in the diff scope — nothing inspected." {
		t.Fatalf("expected the nothing-inspected message, got %q", r.Message)
	}
	if contains(r.Message, "within") || contains(r.Message, "under") {
		t.Fatalf("the skip message must not claim files were within the cap: %q", r.Message)
	}
}

// The mirror direction: a NON-empty filter with a clean file still passes, and
// the pass message names the cap and the exception mechanism.
func TestCheckFileSize_NonEmptyFilterCleanFileStillPasses(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("src", 0755)                                  //nolint:errcheck
	os.WriteFile("src/small.go", []byte("package a\n"), 0644) //nolint:errcheck
	r := checkFileSize(&config.Config{}, gitdiff.NewSet([]string{"src/small.go"}))
	if r.Status != Pass {
		t.Fatalf("clean in-scope file must pass, got %v: %v", r.Status, r.Details)
	}
	if !contains(r.Message, "100-line cap") || !contains(r.Message, "gates.file_size_exceptions") {
		t.Fatalf("pass message must name the cap and the exception mechanism, got %q", r.Message)
	}
}

func TestCheckFileSize_FilterRestrictsScannedSet(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("src", 0755) //nolint:errcheck
	big := makeBigSource(101)
	os.WriteFile("src/included.go", []byte(big), 0644) //nolint:errcheck
	os.WriteFile("src/ignored.go", []byte(big), 0644)  //nolint:errcheck

	r := checkFileSize(&config.Config{}, gitdiff.NewSet([]string{"src/included.go"}))
	if r.Status != Fail {
		t.Fatalf("expected fail for included file, got %v", r.Status)
	}
	if len(r.Details) != 1 || !contains(r.Details[0], "included.go") {
		t.Fatalf("expected only included.go in details, got %v", r.Details)
	}
	for _, d := range r.Details {
		if contains(d, "ignored.go") {
			t.Fatalf("ignored.go must not be flagged when outside filter")
		}
	}
}

func makeBigSource(lines int) string {
	out := ""
	for i := 0; i < lines; i++ {
		out += "x\n"
	}
	return out
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
