package selfupdate

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"
)

// newUpdater builds an Updater pointed at the fake server with a temp HOME/XDG
// dir and a temp binary as the replace target, so no real binary is touched.
func newUpdater(t *testing.T, fs *fakeServer, version string) (*Updater, string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(dir, "cache"))
	bin := filepath.Join(dir, "centinela")
	if err := os.WriteFile(bin, []byte("OLD-BINARY"), 0o755); err != nil {
		t.Fatal(err)
	}
	api := ""
	var client Doer = http.DefaultClient
	if fs != nil {
		api = fs.srv.URL
		client = fs.srv.Client()
	}
	return &Updater{
		Version: version, GOOS: runtime.GOOS, GOARCH: runtime.GOARCH,
		APIBase: api, HTTP: client, TTL: 24 * time.Hour, Now: time.Now,
		Target: func() (string, error) { return bin, nil },
	}, bin
}

// writeCacheFile seeds the cache file with tag at the given age.
func writeCacheFile(t *testing.T, u *Updater, tag string, age time.Duration) {
	t.Helper()
	u.Now = func() time.Time { return time.Unix(1_700_000_000, 0) }
	at := u.Now().Add(-age).Unix()
	p := cachePath()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"latestTag":"` + tag + `","checkedAt":` + strconv.FormatInt(at, 10) + `}`
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// errDoer is a transport that always fails, simulating an offline network.
type errDoer struct{}

func (errDoer) Do(*http.Request) (*http.Response, error) { return nil, errors.New("offline") }
