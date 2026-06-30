package acceptance_test

import (
	"os"
	"strings"
	"testing"
	"time"
)

// Acceptance: specs/cli-self-update.feature
// Scenario: Update installs a newer release and prints old and new versions
func TestCliSelfUpdate_UpdateInstallsNewerRelease(t *testing.T) {
	_, u, bin := newAcUpdater(t, "0.37.0", acFakeOpts{tag: "v0.40.2", withAsset: true, goodSum: true})
	msg, err := u.Update()
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if !strings.Contains(msg, "0.37.0 -> 0.40.2") {
		t.Fatalf("msg = %q", msg)
	}
	if got, _ := os.ReadFile(bin); string(got) != "NEW" {
		t.Fatalf("binary not replaced: %q", got)
	}
}

// Acceptance: specs/cli-self-update.feature
// Scenario: Update is a no-op when already on the latest version
func TestCliSelfUpdate_UpdateNoOp(t *testing.T) {
	_, u, bin := newAcUpdater(t, "0.40.2", acFakeOpts{tag: "v0.40.2", withAsset: true, goodSum: true})
	msg, err := u.Update()
	if err != nil || msg != "already up to date" {
		t.Fatalf("msg=%q err=%v", msg, err)
	}
	if got, _ := os.ReadFile(bin); string(got) != "OLD" {
		t.Fatalf("binary changed on no-op: %q", got)
	}
}

// Acceptance: specs/cli-self-update.feature
// Scenario: --check reports a newer version and exits non-zero with zero writes
func TestCliSelfUpdate_CheckBehindReturnsMsg(t *testing.T) {
	_, u, bin := newAcUpdater(t, "0.37.0", acFakeOpts{tag: "v0.40.2"})
	res, err := u.Check()
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if !res.Behind || !strings.Contains(res.Message, "v0.40.2") {
		t.Fatalf("expected behind=true with v0.40.2 in message, got %+v", res)
	}
	if got, _ := os.ReadFile(bin); string(got) != "OLD" {
		t.Fatalf("binary modified by --check: %q", got)
	}
}

// Acceptance: specs/cli-self-update.feature
// Scenario: --check reports already current and exits zero with zero writes
func TestCliSelfUpdate_CheckCurrentReturnsUpToDate(t *testing.T) {
	_, u, bin := newAcUpdater(t, "0.40.2", acFakeOpts{tag: "v0.40.2"})
	res, err := u.Check()
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if res.Behind || !strings.Contains(res.Message, "up to date") {
		t.Fatalf("expected behind=false with up-to-date message, got %+v", res)
	}
	if got, _ := os.ReadFile(bin); string(got) != "OLD" {
		t.Fatalf("binary modified by --check: %q", got)
	}
}

// Acceptance: specs/cli-self-update.feature
// Scenario: --check honors the TTL cache and makes no network call within TTL
func TestCliSelfUpdate_CheckHonorsTTLCache(t *testing.T) {
	srv, u, _ := newAcUpdater(t, "0.37.0", acFakeOpts{tag: "v0.40.2"})
	seedCache(t, u, "v0.40.2", time.Hour) // fresh: 1h < 24h TTL
	res, err := u.Check()
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if !res.Behind {
		t.Fatalf("expected behind=true from cache, got %+v", res)
	}
	if srv.hits != 0 {
		t.Fatalf("expected zero network calls (cache hit), got %d", srv.hits)
	}
}
