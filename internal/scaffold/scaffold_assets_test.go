package scaffold

import "testing"

func TestReadAssetAndListAssetFiles(t *testing.T) {
	data, err := ReadAsset("CLAUDE.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("expected CLAUDE.md scaffold asset data")
	}
	paths, err := ListAssetFiles("docs/architecture/", ".md")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("expected architecture scaffold markdown assets")
	}
	if _, err := ReadAsset("missing/file.md"); err == nil {
		t.Fatal("expected missing asset error")
	}
}
