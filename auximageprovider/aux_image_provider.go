// auximageprovider exposes an interface for retreiving auxiliary images
// such as watermarks and fallbacks. Default implementation stores those in memory.
package auximageprovider

import (
	"context"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/imagedatanew"
	"github.com/imgproxy/imgproxy/v3/options"
)

// AuxImageProvider is an interface that provides image data and headers based
// on processing options. It is used to retrieve WatermarkImage and FallbackImage.
type AuxImageProvider interface {
	Get(context.Context, *options.ProcessingOptions) (imagedatanew.ImageData, http.Header, error)
}

// memoryAuxImageProvider is a simple implementation of ImageProvider, which returns
// a static saved image data and headers.
type memoryAuxImageProvider struct {
	data    imagedatanew.ImageData
	headers http.Header
}

// newStaticAuxImageProvider creates a new staticImageProvider with the given image data and headers.
func newStaticAuxImageProvider(data imagedatanew.ImageData, headers http.Header) AuxImageProvider {
	return &memoryAuxImageProvider{data: data, headers: headers}
}

// Get returns the static image data and headers stored in the provider.
func (s *memoryAuxImageProvider) Get(_ context.Context, po *options.ProcessingOptions) (imagedatanew.ImageData, http.Header, error) {
	return s.data, s.headers, nil
}
