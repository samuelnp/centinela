package selfupdate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
)

// fakeRelease configures the fake GitHub server's behaviour for one test.
type fakeRelease struct {
	tag       string
	asset     []byte
	withAsset bool
	goodSum   bool
	status    int // when >= 400, every request returns this status
}

type fakeServer struct {
	srv  *httptest.Server
	hits int
}

// newServer stands up an httptest.Server serving releases/latest, the host
// asset, and SHA256SUMS. It counts every request so tests assert exact HTTP
// call counts against the test server (never the real GitHub API).
func newServer(t *testing.T, f fakeRelease) *fakeServer {
	t.Helper()
	fs := &fakeServer{}
	name := AssetName(runtime.GOOS, runtime.GOARCH, f.tag)
	fs.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.hits++
		switch {
		case f.status >= 400:
			w.WriteHeader(f.status)
		case strings.HasSuffix(r.URL.Path, "/releases/latest"):
			writeRelease(w, fs.srv.URL, f, name)
		case strings.HasSuffix(r.URL.Path, "/asset/SHA256SUMS"):
			_, _ = w.Write(sumsFor(name, f.asset, f.goodSum))
		case strings.Contains(r.URL.Path, "/asset/"+name):
			_, _ = w.Write(f.asset)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(fs.srv.Close)
	return fs
}

func writeRelease(w http.ResponseWriter, base string, f fakeRelease, name string) {
	assets := []map[string]string{}
	if f.withAsset {
		assets = append(assets, map[string]string{"name": name, "browser_download_url": base + "/asset/" + name})
	}
	assets = append(assets, map[string]string{"name": "SHA256SUMS", "browser_download_url": base + "/asset/SHA256SUMS"})
	_ = json.NewEncoder(w).Encode(map[string]any{"tag_name": f.tag, "assets": assets})
}

func sumsFor(name string, data []byte, good bool) []byte {
	sum := sha256.Sum256(data)
	h := hex.EncodeToString(sum[:])
	if !good {
		h = strings.Repeat("0", 64)
	}
	return []byte(h + "  " + name + "\n")
}
