package docstring

import "testing"

type stubScanner struct{ called bool }

func (s *stubScanner) Scan([]string, Options) (Report, error) {
	s.called = true
	return Report{Files: 7}, nil
}

func TestFor_ReturnsTheBuiltInGoScanner(t *testing.T) {
	if _, ok := For(GoLang); !ok {
		t.Fatal("the go scanner must be registered by init")
	}
	if _, ok := For("cobol"); ok {
		t.Fatal("an unregistered language must not resolve")
	}
}

func TestRegister_BindsAndRestoresALanguage(t *testing.T) {
	stub := &stubScanner{}
	Register("stub", stub)
	t.Cleanup(func() { delete(scanners, "stub") })

	s, ok := For("stub")
	if !ok {
		t.Fatal("registered scanner must resolve")
	}
	rep, err := s.Scan(nil, Options{})
	if err != nil || !stub.called || rep.Files != 7 {
		t.Fatalf("rep=%+v err=%v called=%v", rep, err, stub.called)
	}
}

func TestUnregister_DropsALanguageBinding(t *testing.T) {
	Register("stub", &stubScanner{})
	Unregister("stub")
	if _, ok := For("stub"); ok {
		t.Fatal("an unregistered language must not resolve")
	}
}
