package integration_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/selfupdate"
)

type integSrv struct {
	*httptest.Server
	hits int
}

func newIntegSrv(t *testing.T, tag string, withAsset bool) *integSrv {
	t.Helper()
	is := &integSrv{}
	name := selfupdate.AssetName(runtime.GOOS, runtime.GOARCH, tag)
	is.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		is.hits++
		assets := []map[string]string{{"name": "SHA256SUMS", "browser_download_url": is.URL + "/sums"}}
		if withAsset {
			assets = append(assets, map[string]string{"name": name, "browser_download_url": is.URL + "/bin"})
		}
		switch {
		case strings.HasSuffix(r.URL.Path, "/releases/latest"):
			_ = json.NewEncoder(w).Encode(map[string]any{"tag_name": tag, "assets": assets})
		case r.URL.Path == "/bin":
			_, _ = w.Write([]byte("NEW"))
		case r.URL.Path == "/sums":
			sum := sha256.Sum256([]byte("NEW"))
			_, _ = w.Write([]byte(hex.EncodeToString(sum[:]) + "  " + name + "\n"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(is.Close)
	return is
}

func newIntegUpdater(t *testing.T, is *integSrv, version string) (*selfupdate.Updater, string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", filepath.Join(dir, "cache"))
	bin := filepath.Join(dir, "centinela")
	if err := os.WriteFile(bin, []byte("OLD"), 0o755); err != nil {
		t.Fatal(err)
	}
	return &selfupdate.Updater{Version: version, GOOS: runtime.GOOS, GOARCH: runtime.GOARCH,
		APIBase: is.URL, HTTP: is.Client(), TTL: 24 * time.Hour, Now: time.Now,
		Target: func() (string, error) { return bin, nil }}, bin
}

// TestInteg_UpdateThenNoOpOnSecondRun verifies Update installs then no-ops.
func TestInteg_UpdateThenNoOpOnSecondRun(t *testing.T) {
	is := newIntegSrv(t, "v0.40.2", true)
	u, bin := newIntegUpdater(t, is, "0.37.0")
	msg, err := u.Update()
	if err != nil || !strings.Contains(msg, "0.37.0 -> 0.40.2") {
		t.Fatalf("first update: msg=%q err=%v", msg, err)
	}
	if got, _ := os.ReadFile(bin); string(got) != "NEW" {
		t.Fatalf("binary not replaced: %q", got)
	}
	u.Version = "0.40.2"
	msg2, err2 := u.Update()
	if err2 != nil || msg2 != "already up to date" {
		t.Fatalf("second update: msg=%q err=%v", msg2, err2)
	}
}

// TestInteg_NoticeCacheInteraction verifies second Notice is cache-served.
func TestInteg_NoticeCacheInteraction(t *testing.T) {
	is := newIntegSrv(t, "v0.40.2", false)
	u, _ := newIntegUpdater(t, is, "0.37.0")
	notice1 := u.Notice()
	if !strings.Contains(notice1, "v0.40.2") {
		t.Fatalf("first notice = %q", notice1)
	}
	hitsAfterFirst := is.hits
	if hitsAfterFirst == 0 {
		t.Fatal("expected at least one HTTP call for first Notice")
	}
	notice2 := u.Notice()
	if !strings.Contains(notice2, "v0.40.2") {
		t.Fatalf("second notice = %q", notice2)
	}
	if is.hits != hitsAfterFirst {
		t.Fatalf("second Notice should use cache: hits went from %d to %d", hitsAfterFirst, is.hits)
	}
}
