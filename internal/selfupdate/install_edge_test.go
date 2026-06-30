package selfupdate

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
)

// installUpdater builds an Updater + Release whose asset/sums URLs point at srv.
func installUpdater(t *testing.T, srv *httptest.Server, assetPath, sumsPath string) (*Updater, *Release) {
	t.Helper()
	u := &Updater{GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, HTTP: srv.Client(),
		Target: func() (string, error) { return filepath.Join(t.TempDir(), "x"), nil }}
	rel := &Release{Tag: "v1", Assets: []asset{
		{Name: u.assetName("v1"), URL: srv.URL + assetPath},
		{Name: sumsAsset, URL: srv.URL + sumsPath},
	}}
	return u, rel
}

func newInstallServer(t *testing.T, sumName string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/asset/bin", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("NEW")) })
	mux.HandleFunc("/asset/SHA256SUMS", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(sumsFor(sumName, []byte("NEW"), true))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func wantKind(t *testing.T, err error, kind Kind) {
	t.Helper()
	var e *Error
	if !errors.As(err, &e) || e.Kind != kind {
		t.Fatalf("want %s error, got %v", kind, err)
	}
}

func TestInstallChecksumNoEntry(t *testing.T) {
	srv := newInstallServer(t, "some-other-name")
	u, rel := installUpdater(t, srv, "/asset/bin", "/asset/SHA256SUMS")
	wantKind(t, u.install(rel), KindChecksum)
}

func TestInstallAssetFetchError(t *testing.T) {
	srv := newInstallServer(t, "ignored")
	u, rel := installUpdater(t, srv, "/asset/gone", "/asset/SHA256SUMS")
	wantKind(t, u.install(rel), KindAPI)
}

func TestInstallSumsFetchError(t *testing.T) {
	srv := newInstallServer(t, "ignored")
	u, rel := installUpdater(t, srv, "/asset/bin", "/asset/gone")
	wantKind(t, u.install(rel), KindAPI)
}

func TestResolveLatestDecodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("{not json"))
	}))
	t.Cleanup(srv.Close)
	u := &Updater{APIBase: srv.URL, HTTP: srv.Client()}
	wantKind(t, func() error { _, e := u.resolveLatest(); return e }(), KindAPI)
}
