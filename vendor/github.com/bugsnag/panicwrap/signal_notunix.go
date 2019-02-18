// +build plan9 windows

package panicwrap

import (
	"os"
)

var signalsToIgnore = []os.Signal{os.Interrupt}
