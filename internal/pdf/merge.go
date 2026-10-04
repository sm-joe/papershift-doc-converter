package pdf

import (
	"context"
	"fmt"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

func Merge(
	ctx context.Context,
	inputs []string,
	output string,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if len(inputs) < 2 {
		return fmt.Errorf("merge requires at least two PDF files")
	}

	if err := api.MergeCreateFile(
		ctx,
		inputs,
		output,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("merge PDFs: %w", err)
	}

	return nil
}