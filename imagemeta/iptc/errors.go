package iptc

import (
	"fmt"

	"github.com/imgproxy/imgproxy/v3/ierrors"
)

type IptcError string

func newIptcError(format string, args ...interface{}) error {
	return ierrors.Wrap(
		IptcError(fmt.Sprintf(format, args...)),
		1,
		ierrors.WithStatusCode(422),
		ierrors.WithPublicMessage("Invalid IPTC data"),
		ierrors.WithShouldReport(false),
	)
}

func (e IptcError) Error() string { return string(e) }
