package main

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

// fakeUpdater builds an httptest-backed Updater + temp binary and wires it into
// the newSelfUpdater seam, so the command paths run fully offline.
func fakeUpdater(t *testing.T, version string, status int) (string, func()) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", filepath.Join(dir, "cache"))
	bin := filepath.Join(dir, "centinela")
	_ = os.WriteFile(bin, []byte("OLD"), 0o755)
	name := selfupdate.AssetName(runtime.GOOS, runtime.GOARCH, "v0.40.2")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case status >= 400:
			w.WriteHeader(status)
		case strings.HasSuffix(r.URL.Path, "/releases/latest"):
			base := "http://" + r.Host
			a := []map[string]string{
				{"name": name, "browser_download_url": base + "/a/" + name},
				{"name": "SHA256SUMS", "browser_download_url": base + "/a/SHA256SUMS"},
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"tag_name": "v0.40.2", "assets": a})
		case strings.HasSuffix(r.URL.Path, "/SHA256SUMS"):
			sum := sha256.Sum256([]byte("NEW"))
			_, _ = w.Write([]byte(hex.EncodeToString(sum[:]) + "  " + name + "\n"))
		default:
			_, _ = w.Write([]byte("NEW"))
		}
	}))
	orig := newSelfUpdater
	newSelfUpdater = func(string) *selfupdate.Updater {
		return &selfupdate.Updater{Version: version, GOOS: runtime.GOOS, GOARCH: runtime.GOARCH,
			APIBase: srv.URL, HTTP: srv.Client(), Now: time.Now, Target: func() (string, error) { return bin, nil }}
	}
	return bin, func() { newSelfUpdater = orig; srv.Close() }
}
