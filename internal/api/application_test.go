package api

import "testing"

func TestNewApplication(t *testing.T) {
	app := NewApplication(t.TempDir())

	if app == nil {
		t.Fatal("expected application")
	}

	if app.Converters == nil {
		t.Fatal("expected converter registry")
	}

	if app.Jobs == nil {
		t.Fatal("expected job service")
	}

	if len(app.Formats) == 0 {
		t.Fatal("expected registered formats")
	}

	converters := app.Converters.All()

	if len(converters) != 1 {
		t.Fatalf(
			"expected 1 converter, got %d",
			len(converters),
		)
	}

	if converters[0].Name() != "libreoffice" {
		t.Fatalf(
			"expected libreoffice, got %s",
			converters[0].Name(),
		)
	}
}
