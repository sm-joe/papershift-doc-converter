package detection

import (
	"archive/zip"
	"bufio"
	"bytes"
	"errors"
	"io"
	"path/filepath"
	"strings"

	"github.com/sm-joe/papershift-doc-converter/internal/formats"
)

const sniffSize = 512

var ErrUnknownFormat = errors.New("unknown file format")

type Detector struct{}

func New() *Detector {
	return &Detector{}
}

func (d *Detector) Detect(name string, r io.Reader) (formats.Format, error) {
	data, err := readSniffBytes(r)
	if err != nil {
		return formats.Format{}, err
	}

	if format, ok := detectBySignature(data); ok {
		return format, nil
	}

	if isZIPSignature(data) {
		if format, ok := detectZipContainer(r); ok {
			return format, nil
		}
	}

	if format, ok := detectText(name, data); ok {
		return format, nil
	}

	if format, ok := detectByExtension(name); ok {
		return format, nil
	}

	return formats.Format{}, ErrUnknownFormat
}

func readSniffBytes(r io.Reader) ([]byte, error) {
	reader := bufio.NewReader(io.LimitReader(r, sniffSize))

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func detectBySignature(data []byte) (formats.Format, bool) {
	if bytes.HasPrefix(data, []byte("%PDF-")) {
		format, ok := formats.Get("pdf")
		return format, ok
	}

	if bytes.HasPrefix(data, []byte("\x89PNG\r\n\x1a\n")) {
		format, ok := formats.Get("png")
		return format, ok
	}

	if bytes.HasPrefix(data, []byte("\xff\xd8\xff")) {
		format, ok := formats.Get("jpg")
		return format, ok
	}

	if bytes.HasPrefix(data, []byte("GIF87a")) ||
		bytes.HasPrefix(data, []byte("GIF89a")) {
		format, ok := formats.Get("gif")
		return format, ok
	}

	if bytes.HasPrefix(data, []byte("BM")) {
		format, ok := formats.Get("bmp")
		return format, ok
	}

	if bytes.HasPrefix(data, []byte("II*\x00")) ||
		bytes.HasPrefix(data, []byte("MM\x00*")) {
		format, ok := formats.Get("tiff")
		return format, ok
	}

	return formats.Format{}, false
}

func isZIPSignature(data []byte) bool {
	return bytes.HasPrefix(data, []byte("PK\x03\x04")) ||
		bytes.HasPrefix(data, []byte("PK\x05\x06")) ||
		bytes.HasPrefix(data, []byte("PK\x07\x08"))
}

func detectZipContainer(r io.Reader) (formats.Format, bool) {
	readerAt, ok := r.(io.ReaderAt)
	if !ok {
		return formats.Format{}, false
	}

	seeker, ok := r.(io.Seeker)
	if !ok {
		return formats.Format{}, false
	}

	currentOffset, err := seeker.Seek(0, io.SeekCurrent)
	if err != nil {
		return formats.Format{}, false
	}

	size, err := seeker.Seek(0, io.SeekEnd)
	if err != nil {
		return formats.Format{}, false
	}

	if _, err := seeker.Seek(currentOffset, io.SeekStart); err != nil {
		return formats.Format{}, false
	}

	archive, err := zip.NewReader(readerAt, size)
	if err != nil {
		return formats.Format{}, false
	}

	var hasContentTypes bool
	var hasWord bool
	var hasExcel bool
	var hasPowerPoint bool

	for _, file := range archive.File {
		name := strings.ReplaceAll(file.Name, "\\", "/")

		switch name {
		case "[Content_Types].xml":
			hasContentTypes = true
		case "word/document.xml":
			hasWord = true
		case "xl/workbook.xml":
			hasExcel = true
		case "ppt/presentation.xml":
			hasPowerPoint = true
		}
	}

	if !hasContentTypes {
		return formats.Format{}, false
	}

	switch {
	case hasWord:
		format, ok := formats.Get("docx")
		return format, ok

	case hasExcel:
		format, ok := formats.Get("xlsx")
		return format, ok

	case hasPowerPoint:
		format, ok := formats.Get("pptx")
		return format, ok
	}

	return formats.Format{}, false
}

func detectText(name string, data []byte) (formats.Format, bool) {
	if !isLikelyText(data) {
		return formats.Format{}, false
	}

	extension := strings.ToLower(filepath.Ext(name))

	switch extension {
	case ".txt":
		format, ok := formats.Get("txt")
		return format, ok

	case ".md":
		format, ok := formats.Get("md")
		return format, ok

	case ".html", ".htm":
		format, ok := formats.Get("html")
		return format, ok

	case ".csv":
		format, ok := formats.Get("csv")
		return format, ok

	case ".tsv":
		format, ok := formats.Get("tsv")
		return format, ok
	}

	return formats.Format{}, false
}

func isLikelyText(data []byte) bool {
	if len(data) == 0 {
		return true
	}

	return !bytes.Contains(data, []byte{0x00})
}

func detectByExtension(name string) (formats.Format, bool) {
	extension := strings.ToLower(filepath.Ext(name))

	for _, format := range formats.All() {
		if strings.EqualFold(format.Extension, extension) {
			return format, true
		}
	}

	return formats.Format{}, false
}
