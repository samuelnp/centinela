package gates

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func TestFindOversizedFilesFallbackRoot(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	big := ""
	for i := 0; i < 101; i++ {
		big += "x\n"
	}
	os.WriteFile("main.go", []byte(big), 0644) //nolint:errcheck
	v, _ := findOversizedFiles(&config.Config{}, nil)
	if len(v) != 0 {
		t.Fatalf("expected no violations when only fallback root is used, got %v", v)
	}
}
