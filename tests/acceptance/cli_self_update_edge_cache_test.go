package acceptance_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/selfupdate"
)

// Acceptance: specs/cli-self-update.feature
// Scenario: Explicit update returns a clear error when GitHub API is unreachable
func TestCliSelfUpdate_ExplicitUpdateNetworkError(t *testing.T) {
	_, u, bin := newAcUpdater(t, "0.37.0", acFakeOpts{tag: "v0.40.2"})
	u.HTTP = acErrDoer{}
	_, err := u.Update()
	if err == nil {
		t.Fatal("expected error on network failure")
	}
	if !strings.Contains(err.Error(), "centinela update") {
		t.Fatalf("error missing context: %q", err.Error())
	}
	if got, _ := os.ReadFile(bin); string(got) != "OLD" {
		t.Fatalf("binary modified on network error: %q", got)
	}
}

// Acceptance: specs/cli-self-update.feature
// Scenario: Stale cache older than the TTL triggers a fresh network check
func TestCliSelfUpdate_StaleCacheRefreshes(t *testing.T) {
	srv, u, _ := newAcUpdater(t, "0.37.0", acFakeOpts{tag: "v0.40.2"})
	seedCache(t, u, "v0.39.0", 48*time.Hour) // stale: 48h > 24h TTL
	notice := u.Notice()
	if !strings.Contains(notice, "v0.40.2") {
		t.Fatalf("notice = %q", notice)
	}
	if srv.hits != 1 {
		t.Fatalf("expected 1 HTTP call for stale cache, got %d", srv.hits)
	}
	data, _ := os.ReadFile(acCachePath())
	if !strings.Contains(string(data), "v0.40.2") {
		t.Fatalf("cache not rewritten with latest tag: %s", data)
	}
}

// Acceptance: specs/cli-self-update.feature
// Scenario: Corrupt or empty cache file triggers a fresh network check without panic
func TestCliSelfUpdate_CorruptCacheRefreshes(t *testing.T) {
	srv, u, _ := newAcUpdater(t, "0.37.0", acFakeOpts{tag: "v0.40.2"})
	p := acCachePath()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("{bad json"), 0o644); err != nil {
		t.Fatal(err)
	}
	notice := u.Notice()
	if !strings.Contains(notice, "v0.40.2") {
		t.Fatalf("notice = %q", notice)
	}
	if srv.hits != 1 {
		t.Fatalf("expected 1 HTTP call after corrupt cache, got %d", srv.hits)
	}
}

// Acceptance: specs/cli-self-update.feature
// Scenario: GitHub API 429 during startup notice fails silently
func TestCliSelfUpdate_RateLimit429SilentOnNotice(t *testing.T) {
	_, u, bin := newAcUpdater(t, "0.37.0", acFakeOpts{status: 429})
	notice := u.Notice()
	if notice != "" {
		t.Fatalf("expected no notice on 429, got %q", notice)
	}
	if got, _ := os.ReadFile(bin); string(got) != "OLD" {
		t.Fatalf("binary modified on 429: %q", got)
	}
}

// Acceptance: specs/cli-self-update.feature
// Scenario: GitHub API 403 during explicit update returns a clear typed error
func TestCliSelfUpdate_RateLimit403ExplicitError(t *testing.T) {
	_, u, bin := newAcUpdater(t, "0.37.0", acFakeOpts{status: 403})
	_, err := u.Update()
	if err == nil {
		t.Fatal("expected error on 403")
	}
	var e *selfupdate.Error
	if !errors.As(err, &e) || e.Kind != selfupdate.KindAPI {
		t.Fatalf("want KindAPI error, got %v (type %T)", err, err)
	}
	if got, _ := os.ReadFile(bin); string(got) != "OLD" {
		t.Fatalf("binary modified on 403: %q", got)
	}
}
