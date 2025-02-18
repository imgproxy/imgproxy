package processing

import (
	"fmt"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagetype"
)

type (
	SaveFormatError string
)

func newSaveFormatError(format imagetype.Type) error {
	return ierrors.Wrap(
		SaveFormatError(fmt.Sprintf("Can't save %s, probably not supported by your libvips", format)),
		1,
		ierrors.WithStatusCode(http.StatusUnprocessableEntity),
		ierrors.WithPublicMessage("Invalid URL"),
		ierrors.WithShouldReport(false),
	)
}

func (e SaveFormatError) Error() string { return string(e) }
