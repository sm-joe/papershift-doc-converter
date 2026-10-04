package pdf

import (
	"context"
	"fmt"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

func Rotate(
	ctx context.Context,
	input string,
	output string,
	rotation int,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	switch rotation {
	case 90, 180, 270:
	default:
		return fmt.Errorf(
			"rotation must be 90, 180, or 270 degrees",
		)
	}

	if err := api.RotateFile(
		ctx,
		input,
		output,
		rotation,
		nil,
		nil,
	); err != nil {
		return fmt.Errorf("rotate PDF: %w", err)
	}

	return nil
}