package errformat

import (
	"fmt"
	"reflect"

	"github.com/sirupsen/logrus"

	"github.com/imgproxy/imgproxy/v3/ierrors"
)

func FormatErrType(errType string, err error) string {
	errType += "_error"

	if _, ok := err.(*ierrors.Error); !ok {
		errType = fmt.Sprintf("%s (%s)", errType, reflect.TypeOf(err).String())
		logrus.Warnf("ErrType: %s", errType)
	}

	return errType
}
