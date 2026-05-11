package gates

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func TestCheckFileSize_AllowsJustifiedException(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("internal", 0755) //nolint:errcheck
	big := ""
	for i := 0; i < 120; i++ {
		big += "x\n"
	}
	os.WriteFile("internal/config_blob.go", []byte(big), 0644) //nolint:errcheck
	cfg := &config.Config{Gates: config.GatesConfig{FileSizeExceptions: []config.FileSizeException{{Path: "internal/config_blob.go", Kind: "configuration", Reason: "large static map", MaxLines: 130}}}}
	r := checkFileSize(cfg, nil)
	if r.Status != Pass || len(r.Details) != 1 {
		t.Fatalf("expected pass with justified detail, got %+v", r)
	}
}

func TestCheckFileSize_FailsWhenExceptionMaxExceeded(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("internal", 0755) //nolint:errcheck
	big := ""
	for i := 0; i < 121; i++ {
		big += "x\n"
	}
	os.WriteFile("internal/domain.go", []byte(big), 0644) //nolint:errcheck
	cfg := &config.Config{Gates: config.GatesConfig{FileSizeExceptions: []config.FileSizeException{{Path: "internal/domain.go", Kind: "domain_atomic", Reason: "single aggregate", MaxLines: 110}}}}
	r := checkFileSize(cfg, nil)
	if r.Status != Fail || len(r.Details) == 0 {
		t.Fatalf("expected fail with details, got %+v", r)
	}
}

func TestFileSizeExceptionMapHandlesNilConfig(t *testing.T) {
	if m := fileSizeExceptionMap(nil); len(m) != 0 {
		t.Fatalf("expected empty map for nil config, got %d", len(m))
	}
}
