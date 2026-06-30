package selfupdate

import "testing"

func TestAssetName(t *testing.T) {
	cases := []struct{ goos, goarch, tag, want string }{
		{"linux", "amd64", "v0.40.2", "centinela-v0.40.2-linux-amd64"},
		{"darwin", "arm64", "v0.40.2", "centinela-v0.40.2-darwin-arm64"},
		{"windows", "amd64", "v0.40.2", "centinela-v0.40.2-windows-amd64.exe"},
	}
	for _, c := range cases {
		if got := AssetName(c.goos, c.goarch, c.tag); got != c.want {
			t.Errorf("AssetName(%s,%s,%s)=%q want %q", c.goos, c.goarch, c.tag, got, c.want)
		}
	}
}

func TestNormalizeAndBehind(t *testing.T) {
	if normalize("v0.40.2") != "0.40.2" || normalize("0.40.2") != "0.40.2" {
		t.Fatal("normalize should strip a single leading v")
	}
	u := &Updater{Version: "0.40.2"}
	if u.behind("v0.40.2") {
		t.Fatal("equal versions must not be behind")
	}
	u.Version = "0.37.0"
	if !u.behind("v0.40.2") {
		t.Fatal("older version must be behind")
	}
}

func TestSumForAndVerify(t *testing.T) {
	data := []byte("payload")
	sums := sumsFor("centinela-x", data, true)
	h, ok := sumFor(sums, "centinela-x")
	if !ok || !verify(data, h) {
		t.Fatalf("sumFor/verify failed ok=%v", ok)
	}
	if _, ok := sumFor(sums, "missing"); ok {
		t.Fatal("unexpected checksum match")
	}
	if verify(data, "deadbeef") {
		t.Fatal("verify must fail on wrong digest")
	}
}

func TestErrorTypeWrapsAndFormats(t *testing.T) {
	e := newErr(KindChecksum, "bad", nil)
	if e.Kind != KindChecksum || e.Unwrap() != nil {
		t.Fatal("error fields")
	}
	if got := e.Error(); got == "" {
		t.Fatal("empty error string")
	}
}
