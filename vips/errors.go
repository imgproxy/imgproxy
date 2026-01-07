package vips

import (
	"net/http"
	"regexp"

	"github.com/imgproxy/imgproxy/v3/errctx"
)

var badImageErrRe = []*regexp.Regexp{
	regexp.MustCompile(`^(\S+)load_source: `),
	regexp.MustCompile(`^(\S+)2vips: `),
	regexp.MustCompile(`^VipsJpeg: `),
	regexp.MustCompile(`XML parse error: `),
}

type VipsError struct{ *errctx.TextError }

func newVipsError(msg string) error {
	var opts []errctx.Option

	for _, re := range badImageErrRe {
		if re.MatchString(msg) {
			opts = []errctx.Option{
				errctx.WithStatusCode(http.StatusUnprocessableEntity),
				errctx.WithPublicMessage("Broken or unsupported image"),
			}
			break
		}
	}

	return VipsError{errctx.NewTextError(msg, 1, opts...)}
}
