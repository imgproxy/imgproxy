// Temporary methods to convert old ImageData to new ImageData and vice versa
package imagedata

import (
	"io"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/imagedatanew"
	"github.com/imgproxy/imgproxy/v3/security"
)

// Converts an old ImageData to a new ImageData
func To(old *ImageData) imagedatanew.ImageData {
	if old == nil {
		return nil
	}

	headers := make(http.Header)
	for k, v := range old.Headers {
		headers.Add(k, v)
	}

	d, err := imagedatanew.NewFromBytes(
		old.Data, headers, security.DefaultOptions(),
	)

	if err != nil {
		panic(err) // temp method, can happen
	}

	return d
}

// Converts a new ImageData to an old ImageData
func From(n imagedatanew.ImageData) *ImageData {
	if n == nil {
		return nil
	}

	data, err := io.ReadAll(n.Reader())
	if err != nil {
		panic(err) // temp method, can happen
	}

	headers := make(map[string]string)

	//nolint:staticcheck
	for k, v := range n.Headers() {
		headers[k] = v[0]
	}

	return &ImageData{
		Data:    data,
		Type:    n.Format(),
		Headers: headers,
	}
}
