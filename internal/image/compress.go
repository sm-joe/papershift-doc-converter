package image

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func Compress(
	ctx context.Context,
	input string,
	output string,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	extension := strings.ToLower(
		filepath.Ext(input),
	)

	args := []string{
		input,
		"-strip",
	}

	switch extension {
	case ".jpg", ".jpeg":
		args = append(
			args,
			"-quality",
			"82",
		)

	case ".png":
		args = append(
			args,
			"-define",
			"png:compression-level=9",
			"-define",
			"png:compression-filter=5",
			"-define",
			"png:compression-strategy=1",
		)

	case ".webp":
		args = append(
			args,
			"-quality",
			"82",
		)

	default:
		return fmt.Errorf(
			"unsupported image format: %s",
			extension,
		)
	}

	args = append(args, output)

	cmd := exec.CommandContext(
		ctx,
		"convert",
		args...,
	)

	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		return fmt.Errorf(
			"compress image with ImageMagick: %w: %s",
			err,
			string(outputBytes),
		)
	}

	return nil
}
