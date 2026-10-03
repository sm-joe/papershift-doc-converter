package libreoffice

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

func writeInput(path string, input io.Reader) error {
	file, err := os.OpenFile(
		path,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0600,
	)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = io.Copy(file, input)

	return err
}

func copyFile(source string, destination io.Writer) error {
	file, err := os.Open(source)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = io.Copy(destination, file)

	return err
}

func replaceExtension(name, extension string) string {
	base := strings.TrimSuffix(
		filepath.Base(name),
		filepath.Ext(name),
	)

	return base + extension
}
