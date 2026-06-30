package acceptance_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/selfupdate"
)

type acFakeOpts struct {
	tag       string
	withAsset bool
	goodSum   bool
	status    int
}

type acSrv struct {
	*httptest.Server
	hits int
}

func newAcServer(t *testing.T, o acFakeOpts) *acSrv {
	t.Helper()
	as := &acSrv{}
	name := selfupdate.AssetName(runtime.GOOS, runtime.GOARCH, o.tag)
	as.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		as.hits++
		if o.status >= 400 {
			w.WriteHeader(o.status)
			return
		}
		switch {
		case strings.HasSuffix(r.URL.Path, "/releases/latest"):
			assets := []map[string]string{{"name": "SHA256SUMS", "browser_download_url": as.URL + "/sums"}}
			if o.withAsset {
				assets = append(assets, map[string]string{"name": name, "browser_download_url": as.URL + "/bin"})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"tag_name": o.tag, "assets": assets})
		case strings.HasSuffix(r.URL.Path, "/bin"):
			_, _ = w.Write([]byte("NEW"))
		case strings.HasSuffix(r.URL.Path, "/sums"):
			sum := sha256.Sum256([]byte("NEW"))
			h := hex.EncodeToString(sum[:])
			if !o.goodSum {
				h = strings.Repeat("0", 64)
			}
			_, _ = w.Write([]byte(h + "  " + name + "\n"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(as.Close)
	return as
}
