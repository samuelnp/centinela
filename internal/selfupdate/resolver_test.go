package selfupdate

import (
	"errors"
	"testing"
)

func TestTargetPathExecutableError(t *testing.T) {
	orig := osExecutable
	osExecutable = func() (string, error) { return "", errors.New("no exe") }
	defer func() { osExecutable = orig }()
	_, err := targetPath()
	var e *Error
	if !errors.As(err, &e) || e.Kind != KindReplace {
		t.Fatalf("want replace error, got %v", err)
	}
}

func TestTargetPathSymlinkError(t *testing.T) {
	oe, es := osExecutable, evalSymlinks
	osExecutable = func() (string, error) { return "/some/path", nil }
	evalSymlinks = func(string) (string, error) { return "", errors.New("broken link") }
	defer func() { osExecutable, evalSymlinks = oe, es }()
	_, err := targetPath()
	var e *Error
	if !errors.As(err, &e) || e.Kind != KindReplace {
		t.Fatalf("want replace error, got %v", err)
	}
}

func TestFetchBytesTransportError(t *testing.T) {
	u := &Updater{HTTP: errDoer{}}
	if _, err := u.fetchBytes("http://example.invalid/x"); err == nil {
		t.Fatal("want transport error")
	}
}
