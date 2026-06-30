package acceptance_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/selfupdate"
)

// Acceptance: specs/cli-self-update.feature
// Scenario: Checksum mismatch aborts without touching the installed binary
func TestCliSelfUpdate_ChecksumMismatch(t *testing.T) {
	_, u, bin := newAcUpdater(t, "0.37.0", acFakeOpts{tag: "v0.40.2", withAsset: true, goodSum: false})
	_, err := u.Update()
	var e *selfupdate.Error
	if !errors.As(err, &e) || e.Kind != selfupdate.KindChecksum {
		t.Fatalf("want KindChecksum error, got %v", err)
	}
	if !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("error message missing 'checksum': %q", err.Error())
	}
	if got, _ := os.ReadFile(bin); string(got) != "OLD" {
		t.Fatalf("binary changed despite checksum mismatch: %q", got)
	}
	entries, _ := os.ReadDir(filepath.Dir(bin))
	for _, en := range entries {
		if strings.HasPrefix(en.Name(), ".centinela-update-") {
			t.Fatalf("leftover temp file after checksum failure: %s", en.Name())
		}
	}
}

// Acceptance: specs/cli-self-update.feature
// Scenario: Missing asset for the host platform returns a typed error with no partial write
func TestCliSelfUpdate_MissingAsset(t *testing.T) {
	_, u, bin := newAcUpdater(t, "0.37.0", acFakeOpts{tag: "v0.40.2", withAsset: false})
	_, err := u.Update()
	var e *selfupdate.Error
	if !errors.As(err, &e) || e.Kind != selfupdate.KindPlatform {
		t.Fatalf("want KindPlatform error, got %v", err)
	}
	if got, _ := os.ReadFile(bin); string(got) != "OLD" {
		t.Fatalf("binary changed despite missing asset: %q", got)
	}
	entries, _ := os.ReadDir(filepath.Dir(bin))
	for _, en := range entries {
		if strings.HasPrefix(en.Name(), ".centinela-update-") {
			t.Fatalf("unexpected temp file on missing asset: %s", en.Name())
		}
	}
}

// Acceptance: specs/cli-self-update.feature
// Scenario: Unwritable install directory returns a typed error and leaves binary untouched
func TestCliSelfUpdate_PermissionDenied(t *testing.T) {
	_, u, bin := newAcUpdater(t, "0.37.0", acFakeOpts{tag: "v0.40.2", withAsset: true, goodSum: true})
	dir := filepath.Dir(bin)
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(dir, 0o755) //nolint:errcheck
	_, err := u.Update()
	var e *selfupdate.Error
	if !errors.As(err, &e) || e.Kind != selfupdate.KindPermission {
		t.Fatalf("want KindPermission error, got %v", err)
	}
	if err := os.Chmod(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(bin); string(got) != "OLD" {
		t.Fatalf("binary changed despite permission failure: %q", got)
	}
}
