package acceptance_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/selfupdate"
)

// Acceptance: specs/cli-self-update.feature
// Scenario: All network calls target the httptest.Server and not the real GitHub API
func TestCliSelfUpdate_NetworkIsolation(t *testing.T) {
	srv, u, _ := newAcUpdater(t, "0.37.0", acFakeOpts{tag: "v0.40.2", withAsset: true, goodSum: true})
	if _, err := u.Update(); err != nil {
		t.Fatalf("update: %v", err)
	}
	if srv.hits == 0 {
		t.Fatal("expected HTTP hits to test server")
	}
	if !strings.HasPrefix(u.APIBase, "http://127.0.0.1") {
		t.Fatalf("APIBase not isolated: %s", u.APIBase)
	}
}

// Acceptance: specs/cli-self-update.feature
// Scenario: Version comparison strips leading v from the release tag
func TestCliSelfUpdate_VersionNormalizationEqual(t *testing.T) {
	_, u, bin := newAcUpdater(t, "0.40.2", acFakeOpts{tag: "v0.40.2"})
	res, err := u.Check()
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if res.Behind {
		t.Fatalf("v0.40.2 vs 0.40.2 should be treated as equal, got behind=true")
	}
	if got, _ := os.ReadFile(bin); string(got) != "OLD" {
		t.Fatalf("binary modified on equal version: %q", got)
	}
}

// Acceptance: specs/cli-self-update.feature
// Scenario: Asset name is constructed with the leading v from the tag
func TestCliSelfUpdate_AssetNameWithLeadingV(t *testing.T) {
	name := selfupdate.AssetName("linux", "amd64", "v0.40.2")
	if name != "centinela-v0.40.2-linux-amd64" {
		t.Fatalf("AssetName = %q", name)
	}
}

// Acceptance: specs/cli-self-update.feature
// Scenario: Asset name for Windows carries the .exe suffix
func TestCliSelfUpdate_AssetNameWindowsExe(t *testing.T) {
	name := selfupdate.AssetName("windows", "amd64", "v0.40.2")
	if name != "centinela-v0.40.2-windows-amd64.exe" {
		t.Fatalf("AssetName = %q", name)
	}
}

// Acceptance: specs/cli-self-update.feature
// Scenario: dev build prints an informational message and skips the update
func TestCliSelfUpdate_DevBuildSkipsUpdate(t *testing.T) {
	srv, u, bin := newAcUpdater(t, "dev", acFakeOpts{tag: "v0.40.2", withAsset: true, goodSum: true})
	msg, err := u.Update()
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if !strings.Contains(msg, "development build") {
		t.Fatalf("msg = %q", msg)
	}
	if srv.hits != 0 {
		t.Fatalf("dev build hit network %d times", srv.hits)
	}
	if got, _ := os.ReadFile(bin); string(got) != "OLD" {
		t.Fatal("dev build modified binary")
	}
}

// Acceptance: specs/cli-self-update.feature
// Scenario: dev build suppresses the startup notice
func TestCliSelfUpdate_DevBuildSuppressesNotice(t *testing.T) {
	srv, u, _ := newAcUpdater(t, "dev", acFakeOpts{tag: "v0.40.2"})
	notice := u.Notice()
	if notice != "" {
		t.Fatalf("notice should be empty for dev build, got %q", notice)
	}
	if srv.hits != 0 {
		t.Fatalf("dev build hit network %d times", srv.hits)
	}
}
