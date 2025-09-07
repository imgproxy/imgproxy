package handlers

import (
	"github.com/imgproxy/imgproxy/v3/auximageprovider"
	"github.com/imgproxy/imgproxy/v3/fetcher"
	"github.com/imgproxy/imgproxy/v3/headerwriter"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/semaphores"
)

// Context defines the input interface handler needs to operate.
// In a nutshell, this interface strips ImgProxy definition from implementation.
// All the dependent components could share the same global interface.
//
// It might as well be implemented on the Handler struct itself, no matter.
// However, in this case, we'we got to implement it on every Handler struct.
type Context interface {
	HeaderWriter() *headerwriter.Writer
	Fetcher() *fetcher.Fetcher
	Semaphores() *semaphores.Semaphores
	FallbackImage() auximageprovider.Provider
	WatermarkImage() auximageprovider.Provider
	ImageDataFactory() *imagedata.Factory
}
