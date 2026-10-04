package pdf

import (
	"context"
	"fmt"
	"os/exec"
)

func Compress(
	ctx context.Context,
	input string,
	output string,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	cmd := exec.CommandContext(
		ctx,
		"gs",
		"-sDEVICE=pdfwrite",
		"-dCompatibilityLevel=1.7",
		"-dPDFSETTINGS=/ebook",
		"-dNOPAUSE",
		"-dQUIET",
		"-dBATCH",
		"-sOutputFile="+output,
		input,
	)

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		return fmt.Errorf("compress PDF with Ghostscript: %w", err)
	}

	return nil
}
