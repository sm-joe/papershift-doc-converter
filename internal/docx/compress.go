package docx

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	imagecompress "github.com/sm-joe/papershift-doc-converter/internal/image"
)

func Compress(
	ctx context.Context,
	input string,
	output string,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	reader, err := zip.OpenReader(input)
	if err != nil {
		return fmt.Errorf("open DOCX: %w", err)
	}
	defer func() {
		_ = reader.Close()
	}()

	outputDir := filepath.Dir(output)

	if err := os.MkdirAll(outputDir, 0700); err != nil {
		return fmt.Errorf(
			"create DOCX output directory: %w",
			err,
		)
	}

	tempDir, err := os.MkdirTemp(
		outputDir,
		"docx-compress-*",
	)
	if err != nil {
		return fmt.Errorf(
			"create DOCX compression workspace: %w",
			err,
		)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	file, err := os.Create(output)
	if err != nil {
		return fmt.Errorf(
			"create compressed DOCX: %w",
			err,
		)
	}

	writer := zip.NewWriter(file)

	success := false

	defer func() {
		if !success {
			_ = writer.Close()
			_ = file.Close()
			_ = os.Remove(output)
		}
	}()

	for _, entry := range reader.File {
		if err := ctx.Err(); err != nil {
			return err
		}

		header := entry.FileHeader

		if entry.FileInfo().IsDir() {
			header.Method = zip.Store

			if err := createEmptyEntry(writer, &header); err != nil {
				return err
			}

			continue
		}

		if isCompressibleImage(entry.Name) {
			compressedPath, ok, err := compressImageEntry(
				ctx,
				entry,
				tempDir,
			)
			if err != nil {
				return err
			}

			if ok {
				if err := writeFileEntry(
					writer,
					&header,
					compressedPath,
				); err != nil {
					return err
				}

				continue
			}
		}

		header.Method = zip.Deflate

		entryWriter, err := writer.CreateHeader(&header)
		if err != nil {
			return fmt.Errorf(
				"create DOCX entry %q: %w",
				entry.Name,
				err,
			)
		}

		entryReader, err := entry.Open()
		if err != nil {
			return fmt.Errorf(
				"open DOCX entry %q: %w",
				entry.Name,
				err,
			)
		}

		_, copyErr := io.Copy(entryWriter, entryReader)
		closeErr := entryReader.Close()

		if copyErr != nil {
			return fmt.Errorf(
				"copy DOCX entry %q: %w",
				entry.Name,
				copyErr,
			)
		}

		if closeErr != nil {
			return fmt.Errorf(
				"close DOCX entry %q: %w",
				entry.Name,
				closeErr,
			)
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf(
			"finalize compressed DOCX: %w",
			err,
		)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf(
			"close compressed DOCX: %w",
			err,
		)
	}

	success = true

	return nil
}

func isCompressibleImage(name string) bool {
	name = strings.ToLower(name)

	if !strings.HasPrefix(name, "word/media/") {
		return false
	}

	switch filepath.Ext(name) {
	case ".jpg", ".jpeg", ".png":
		return true
	default:
		return false
	}
}

func compressImageEntry(
	ctx context.Context,
	entry *zip.File,
	tempDir string,
) (string, bool, error) {
	originalPath := filepath.Join(
		tempDir,
		"original"+filepath.Ext(entry.Name),
	)

	compressedPath := filepath.Join(
		tempDir,
		"compressed"+filepath.Ext(entry.Name),
	)

	if err := extractZipEntry(entry, originalPath); err != nil {
		return "", false, err
	}

	if err := imagecompress.Compress(
		ctx,
		originalPath,
		compressedPath,
	); err != nil {
		return "", false, fmt.Errorf(
			"compress DOCX image %q: %w",
			entry.Name,
			err,
		)
	}

	originalInfo, err := os.Stat(originalPath)
	if err != nil {
		return "", false, fmt.Errorf(
			"stat original DOCX image %q: %w",
			entry.Name,
			err,
		)
	}

	compressedInfo, err := os.Stat(compressedPath)
	if err != nil {
		return "", false, fmt.Errorf(
			"stat compressed DOCX image %q: %w",
			entry.Name,
			err,
		)
	}

	if compressedInfo.Size() >= originalInfo.Size() {
		return originalPath, true, nil
	}

	return compressedPath, true, nil
}

func extractZipEntry(
	entry *zip.File,
	output string,
) error {
	reader, err := entry.Open()
	if err != nil {
		return fmt.Errorf(
			"open DOCX image %q: %w",
			entry.Name,
			err,
		)
	}
	defer func() {
		_ = reader.Close()
	}()

	file, err := os.Create(output)
	if err != nil {
		return fmt.Errorf(
			"create temporary DOCX image: %w",
			err,
		)
	}

	_, copyErr := io.Copy(file, reader)
	closeErr := file.Close()

	if copyErr != nil {
		return fmt.Errorf(
			"extract DOCX image %q: %w",
			entry.Name,
			copyErr,
		)
	}

	if closeErr != nil {
		return fmt.Errorf(
			"close temporary DOCX image: %w",
			closeErr,
		)
	}

	return nil
}

func writeFileEntry(
	writer *zip.Writer,
	header *zip.FileHeader,
	path string,
) error {
	header.Method = zip.Store

	entryWriter, err := writer.CreateHeader(header)
	if err != nil {
		return fmt.Errorf(
			"create compressed DOCX image entry %q: %w",
			header.Name,
			err,
		)
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf(
			"open compressed DOCX image %q: %w",
			header.Name,
			err,
		)
	}

	_, copyErr := io.Copy(entryWriter, file)
	closeErr := file.Close()

	if copyErr != nil {
		return fmt.Errorf(
			"write compressed DOCX image %q: %w",
			header.Name,
			copyErr,
		)
	}

	if closeErr != nil {
		return fmt.Errorf(
			"close compressed DOCX image %q: %w",
			header.Name,
			closeErr,
		)
	}

	return nil
}

func createEmptyEntry(
	writer *zip.Writer,
	header *zip.FileHeader,
) error {
	if _, err := writer.CreateHeader(header); err != nil {
		return fmt.Errorf(
			"create DOCX directory entry %q: %w",
			header.Name,
			err,
		)
	}

	return nil
}
