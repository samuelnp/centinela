package acceptance_test

import (
	"os"
	"strings"
	"testing"
	"time"
)

// Acceptance: specs/cli-self-update.feature
// Scenario: Startup notice appears when running an older version and cache is stale
func TestCliSelfUpdate_StartupNoticeStaleCache(t *testing.T) {
	_, u, bin := newAcUpdater(t, "0.37.0", acFakeOpts{tag: "v0.40.2"})
	notice := u.Notice()
	if !strings.Contains(notice, "v0.40.2") {
		t.Fatalf("expected notice mentioning v0.40.2, got %q", notice)
	}
	if got, _ := os.ReadFile(bin); string(got) != "OLD" {
		t.Fatalf("binary modified by notice: %q", got)
	}
	data, err := os.ReadFile(acCachePath())
	if err != nil || !strings.Contains(string(data), "v0.40.2") {
		t.Fatalf("cache not written after stale notice: err=%v data=%s", err, data)
	}
}

// Acceptance: specs/cli-self-update.feature
// Scenario: Startup notice is suppressed when the cache is within the TTL
func TestCliSelfUpdate_StartupNoticeSuppressedWithinTTL(t *testing.T) {
	// "Suppressed" here refers to the network check being suppressed (throttled),
	// not the update-available notice itself. The notice still appears from cache data.
	srv, u, _ := newAcUpdater(t, "0.37.0", acFakeOpts{tag: "v0.40.2"})
	seedCache(t, u, "v0.40.2", time.Hour) // fresh: 1h < 24h TTL
	notice := u.Notice()
	if srv.hits != 0 {
		t.Fatalf("expected zero network calls within TTL, got %d", srv.hits)
	}
	// Notice is shown from cache (0.37.0 < v0.40.2); the cache file is not rewritten.
	if !strings.Contains(notice, "v0.40.2") {
		t.Fatalf("expected notice from cache, got %q", notice)
	}
}

// Acceptance: specs/cli-self-update.feature
// Scenario: Startup notice is suppressed when already on the latest version
func TestCliSelfUpdate_StartupNoticeSuppressedWhenCurrent(t *testing.T) {
	_, u, bin := newAcUpdater(t, "0.40.2", acFakeOpts{tag: "v0.40.2"})
	notice := u.Notice()
	if notice != "" {
		t.Fatalf("no notice expected when current, got %q", notice)
	}
	if got, _ := os.ReadFile(bin); string(got) != "OLD" {
		t.Fatalf("binary modified by notice: %q", got)
	}
}

// Acceptance: specs/cli-self-update.feature
// Scenario: Startup notice fails silently when the GitHub API is unreachable
func TestCliSelfUpdate_StartupNoticeFailsSilent(t *testing.T) {
	_, u, bin := newAcUpdater(t, "0.37.0", acFakeOpts{status: 503})
	notice := u.Notice()
	if notice != "" {
		t.Fatalf("notice should be empty on API error, got %q", notice)
	}
	if got, _ := os.ReadFile(bin); string(got) != "OLD" {
		t.Fatalf("binary modified despite API error: %q", got)
	}
}

// Acceptance: specs/cli-self-update.feature
// Scenario: Startup notice never auto-installs
func TestCliSelfUpdate_StartupNoticeNeverAutoInstalls(t *testing.T) {
	_, u, bin := newAcUpdater(t, "0.37.0", acFakeOpts{tag: "v0.40.2"})
	_ = u.Notice()
	if got, _ := os.ReadFile(bin); string(got) != "OLD" {
		t.Fatalf("notice must never install binary, got: %q", got)
	}
}
