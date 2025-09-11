package generichttp

import (
	"net/http"

	"github.com/imgproxy/imgproxy/v3/ierrors"
)

type (
	SourceAddressError string
)

func newSourceAddressError(msg string) error {
	return ierrors.Wrap(
		SourceAddressError(msg),
		1,
		ierrors.WithStatusCode(http.StatusNotFound),
		ierrors.WithPublicMessage("Invalid source URL"),
		ierrors.WithShouldReport(false),
	)
}

func (e SourceAddressError) Error() string { return string(e) }
