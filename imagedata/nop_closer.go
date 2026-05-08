package imagedata

// nopCloser is a wrapper around ImageData that overrides the Close method to
// do nothing.
// This is useful for cases where we return a shared ImageData that should not
// be closed by the caller.
type nopCloser struct {
	ImageData
}

// NopCloser returns a new ImageData with a no-op Close method that wraps the given ImageData.
func NopCloser(data ImageData) ImageData {
	return &nopCloser{data}
}

// Close does nothing to prevent the wrapped ImageData from being closed.
func (nc *nopCloser) Close() error {
	return nil
}
