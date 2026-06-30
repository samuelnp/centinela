package selfupdate

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateChecksumMismatch(t *testing.T) {
	fs := newServer(t, fakeRelease{tag: "v0.40.2", asset: []byte("NEW"), withAsset: true, goodSum: false})
	u, bin := newUpdater(t, fs, "0.37.0")
	_, err := u.Update()
	var e *Error
	if !errors.As(err, &e) || e.Kind != KindChecksum {
		t.Fatalf("want checksum error, got %v", err)
	}
	if got, _ := os.ReadFile(bin); string(got) != "OLD-BINARY" {
		t.Fatal("binary changed on checksum mismatch")
	}
	assertNoTemp(t, filepath.Dir(bin))
}

func TestUpdateMissingAsset(t *testing.T) {
	fs := newServer(t, fakeRelease{tag: "v0.40.2", withAsset: false, goodSum: true})
	u, bin := newUpdater(t, fs, "0.37.0")
	_, err := u.Update()
	var e *Error
	if !errors.As(err, &e) || e.Kind != KindPlatform {
		t.Fatalf("want platform error, got %v", err)
	}
	assertNoTemp(t, filepath.Dir(bin))
}

func TestUpdatePermissionDenied(t *testing.T) {
	fs := newServer(t, fakeRelease{tag: "v0.40.2", asset: []byte("NEW"), withAsset: true, goodSum: true})
	u, bin := newUpdater(t, fs, "0.37.0")
	dir := filepath.Dir(bin)
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(dir, 0o755) //nolint:errcheck
	_, err := u.Update()
	var e *Error
	if !errors.As(err, &e) || e.Kind != KindPermission {
		t.Fatalf("want permission error, got %v", err)
	}
	if got, _ := os.ReadFile(bin); string(got) != "OLD-BINARY" {
		t.Fatal("binary changed despite permission failure")
	}
}

func TestUpdateAPIError(t *testing.T) {
	fs := newServer(t, fakeRelease{tag: "v0.40.2", status: 403})
	u, _ := newUpdater(t, fs, "0.37.0")
	_, err := u.Update()
	var e *Error
	if !errors.As(err, &e) || e.Kind != KindAPI {
		t.Fatalf("want api error, got %v", err)
	}
}

func assertNoTemp(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".centinela-update-") {
			t.Fatalf("leftover temp file: %s", e.Name())
		}
	}
}
