package imagedata

import "errors"

type httpError interface {
	Timeout() bool
}

func checkTimeoutErr(err error) error {
	if httpErr, ok := err.(httpError); ok && httpErr.Timeout() {
		return errors.New("The image request timed out")
	}
	return err
}
