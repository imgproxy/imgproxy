// auximagedata exposes an interface for retreiving auxiliary images
// such as watermarks and fallbacks. Default implementation stores those in memory.

package auximageprovider

import (
	"context"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
)

// Provider is an interface that provides image data and headers based
// on processing options. It is used to retrieve WatermarkImage and FallbackImage.
type Provider interface {
	Get(context.Context, *options.ProcessingOptions) (imagedata.ImageData, http.Header, error)
}
