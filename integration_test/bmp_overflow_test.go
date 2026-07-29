package integration_test

import (
	"net/http"
)

// TestBMPUnsupportedBitDepthRejected is a regression test for a heap-buffer-overflow
// that used to exist in the custom BMP loader (vips/bmpload.c). For uncompressed
// (BI_RGB) BMP input, the `bpp` field was never validated: bytes_per_pixel was derived
// from any bpp, any bpp >= 24 was routed into the 24/32-bit row generator, and that
// generator read a row sized from the unvalidated bpp into a row buffer allocated for
// a 4-byte-per-pixel maximum. testdata/bmp-invalid-bpp.bmp declares bpp=40 at
// width=512, which used to read a 2560-byte row into a 2052-byte allocation (a
// 508-byte overflow).
func (s *ProcessingHandlerTestSuite) TestBMPUnsupportedBitDepthRejected() {
	res := s.GET("/unsafe/rs:fill:4:4/plain/local:///bmp-invalid-bpp.bmp")
	defer res.Body.Close()

	s.Require().Equal(http.StatusInternalServerError, res.StatusCode)
}
