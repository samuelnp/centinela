package selfupdate

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewWiresHostDefaults(t *testing.T) {
	u := New("1.2.3")
	if u.Version != "1.2.3" || u.HTTP == nil || u.Now == nil || u.Target == nil {
		t.Fatalf("New left a field unset: %+v", u)
	}
	if u.APIBase != defaultAPIBase || u.TTL != defaultTTL {
		t.Fatalf("New defaults wrong: %q %v", u.APIBase, u.TTL)
	}
}

func TestErrorWrappedFormats(t *testing.T) {
	e := newErr(KindReplace, "swap", errors.New("cause"))
	if e.Unwrap() == nil || e.Error() == "" {
		t.Fatal("wrapped error formatting")
	}
}

func TestCachePathHomeFallback(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", "")
	t.Setenv("HOME", "/home/x")
	want := filepath.Join("/home/x", ".cache", "centinela", "update-check.json")
	if got := cachePath(); got != want {
		t.Fatalf("cachePath = %q want %q", got, want)
	}
}

func TestFetchBytesNon200AndBadURL(t *testing.T) {
	fs := newServer(t, fakeRelease{tag: "v1"})
	u, _ := newUpdater(t, fs, "0.1.0")
	if _, err := u.fetchBytes(fs.srv.URL + "/asset/missing"); err == nil {
		t.Fatal("want non-200 error")
	}
	if _, err := u.fetchBytes("://bad-url"); err == nil {
		t.Fatal("want bad-request error")
	}
}

func TestResolveLatestBadURL(t *testing.T) {
	u, _ := newUpdater(t, nil, "0.1.0")
	u.APIBase = "://bad"
	if _, err := u.resolveLatest(); err == nil {
		t.Fatal("want build-request error")
	}
}

func TestInstallMissingSumsAsset(t *testing.T) {
	u, _ := newUpdater(t, nil, "0.1.0")
	rel := &Release{Tag: "v1", Assets: []asset{{
		Name: u.assetName("v1"), URL: "http://example/x",
	}}}
	err := u.install(rel)
	var e *Error
	if !errors.As(err, &e) || e.Kind != KindPlatform {
		t.Fatalf("want platform error, got %v", err)
	}
}

func TestWriteCacheMkdirError(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "blocker")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CACHE_HOME", file) // a file, not a dir → MkdirAll fails
	u := &Updater{Now: func() time.Time { return time.Unix(1, 0) }}
	u.writeCache("v1") // best-effort: must not panic
	if _, ok := u.readCache(); ok {
		t.Fatal("write should have failed silently")
	}
}
