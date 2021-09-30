package processing

import (
	"context"

	"github.com/imgproxy/imgproxy/v3/router"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func copyMemoryAndCheckTimeout(ctx context.Context, img *vips.Image) error {
	err := img.CopyMemory()
	router.CheckTimeout(ctx)
	return err
}
