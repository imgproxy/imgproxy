package imagedatanew

import (
	"io"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/imagemeta"
	"github.com/imgproxy/imgproxy/v3/imagetype"
)

// ImageData is an interface that defines methods for reading image data and metadata
type ImageData interface {
	io.Closer               // Close closes the image data and releases any resources held by it
	Reader() io.ReadSeeker  // Reader returns a new ReadSeeker for the image data
	Meta() imagemeta.Meta   // Meta returns the metadata of the image data
	Format() imagetype.Type // Format returns the image format from the metadata (shortcut)
	Size() (int, error)     // Size returns the size of the image data in bytes

	// This will be removed in the future
	Headers() http.Header // Headers returns the HTTP headers of the image data, will be removed in the future
}
