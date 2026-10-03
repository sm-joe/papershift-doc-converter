package formats

import "testing"

func TestAllReturnsRegisteredFormats(t *testing.T) {
	formats := All()

	if len(formats) == 0 {
		t.Fatal("expected registered formats")
	}
}

func TestGetExistingFormat(t *testing.T) {
	format, ok := Get("docx")

	if !ok {
		t.Fatal("expected docx to be registered")
	}

	if format.Extension != ".docx" {
		t.Fatalf("expected .docx, got %s", format.Extension)
	}

	if format.Family != FamilyDocument {
		t.Fatalf("expected document family, got %s", format.Family)
	}
}

func TestGetUnknownFormat(t *testing.T) {
	_, ok := Get("does-not-exist")

	if ok {
		t.Fatal("expected unknown format lookup to fail")
	}
}
