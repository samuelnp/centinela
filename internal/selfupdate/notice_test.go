package selfupdate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNoticeStaleCacheRefreshes(t *testing.T) {
	fs := newServer(t, fakeRelease{tag: "v0.40.2"})
	u, _ := newUpdater(t, fs, "0.37.0")
	writeCacheFile(t, u, "v0.39.0", 48*time.Hour) // older than TTL
	notice := u.Notice()
	if !strings.Contains(notice, "v0.40.2") || fs.hits != 1 {
		t.Fatalf("notice=%q hits=%d", notice, fs.hits)
	}
	if tag, ok := u.readCache(); !ok || tag != "v0.40.2" {
		t.Fatalf("cache not refreshed: %q ok=%v", tag, ok)
	}
}

func TestNoticeWithinTTLNoNetwork(t *testing.T) {
	fs := newServer(t, fakeRelease{tag: "v0.40.2"})
	u, _ := newUpdater(t, fs, "0.37.0")
	writeCacheFile(t, u, "v0.40.2", time.Hour)
	if notice := u.Notice(); !strings.Contains(notice, "v0.40.2") || fs.hits != 0 {
		t.Fatalf("notice=%q hits=%d", notice, fs.hits)
	}
}

func TestNoticeSuppressedWhenCurrent(t *testing.T) {
	fs := newServer(t, fakeRelease{tag: "v0.40.2"})
	u, _ := newUpdater(t, fs, "0.40.2")
	if notice := u.Notice(); notice != "" {
		t.Fatalf("expected no notice, got %q", notice)
	}
}

func TestNoticeFailsSilentOnError(t *testing.T) {
	fs := newServer(t, fakeRelease{tag: "v0.40.2", status: 429})
	u, _ := newUpdater(t, fs, "0.37.0")
	if notice := u.Notice(); notice != "" {
		t.Fatalf("expected silent, got %q", notice)
	}
}

func TestNoticeDevSuppressed(t *testing.T) {
	fs := newServer(t, fakeRelease{tag: "v0.40.2"})
	u, _ := newUpdater(t, fs, "dev")
	if notice := u.Notice(); notice != "" || fs.hits != 0 {
		t.Fatalf("notice=%q hits=%d", notice, fs.hits)
	}
}

func TestNoticeCorruptCacheRefreshes(t *testing.T) {
	fs := newServer(t, fakeRelease{tag: "v0.40.2"})
	u, _ := newUpdater(t, fs, "0.37.0")
	if err := os.MkdirAll(filepath.Dir(cachePath()), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cachePath(), []byte("{bad json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if notice := u.Notice(); !strings.Contains(notice, "v0.40.2") || fs.hits != 1 {
		t.Fatalf("notice=%q hits=%d", notice, fs.hits)
	}
}
