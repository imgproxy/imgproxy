// imagedata provides shared ImageData interface for working with image data.
package imagedatanew

import (
	"bytes"
	"io"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/asyncbuffer"
	"github.com/imgproxy/imgproxy/v3/imagemeta"
	"github.com/imgproxy/imgproxy/v3/imagetype"
)

// ImageData is an interface that defines methods for reading image data and metadata
type ImageData interface {
	io.Closer               // Close closes the image data and releases any resources held by it
	Reader() io.ReadSeeker  // Reader returns a new ReadSeeker for the image data
	Meta() imagemeta.Meta   // Meta returns the metadata of the image data
	Format() imagetype.Type // Format returns the image format from the metadata (shortcut)

	// Will be removed from the interface in the future (DEPRECATED)
	Headers() http.Header // Headers returns the HTTP headers of the image data, if applicable
}

// imageDataResponse is a struct that implements the ImageData interface for http.Response
type imageDataResponse struct {
	b       *asyncbuffer.AsyncBuffer // AsyncBuffer instance
	c       io.Closer                // Closer for the original response body
	meta    imagemeta.Meta           // Metadata of the image data
	headers http.Header              // Headers for the response, if applicable
}

// imageDataBytes is a struct that implements the ImageData interface for a byte slice
type imageDataBytes struct {
	b       []byte         // ReadSeeker for the image data
	meta    imagemeta.Meta // Metadata of the image data
	headers http.Header    // Headers for the response, if applicable
}

// Reader returns a ReadSeeker for the image data
func (r *imageDataResponse) Reader() io.ReadSeeker {
	return r.b.Reader()
}

// Close closes the response body (hence, response) and the async buffer itself
func (r *imageDataResponse) Close() error {
	if r.c != nil {
		defer r.c.Close()
	}

	return r.b.Close()
}

// Meta returns the metadata of the image data
func (r *imageDataResponse) Meta() imagemeta.Meta {
	return r.meta
}

// Format returns the image format from the metadata
func (r *imageDataResponse) Format() imagetype.Type {
	return r.meta.Format()
}

// Headers returns the headers of the image data, if applicable
func (r *imageDataResponse) Headers() http.Header {
	return r.headers
}

// Reader returns a ReadSeeker for the image data
func (b *imageDataBytes) Reader() io.ReadSeeker {
	return bytes.NewReader(b.b)
}

// Close does nothing for imageDataBytes as it does not hold any resources
func (b *imageDataBytes) Close() error {
	return nil
}

// Meta returns the metadata of the image data
func (b *imageDataBytes) Meta() imagemeta.Meta {
	return b.meta
}

// Format returns the image format from the metadata
func (r *imageDataBytes) Format() imagetype.Type {
	return r.meta.Format()
}

// Headers returns the headers of the image data, if applicable
func (r *imageDataBytes) Headers() http.Header {
	return r.headers
}
