package gliblog

/*
#cgo pkg-config: glib-2.0
#include "gliblog.h"
*/
import "C"
import log "github.com/sirupsen/logrus"

//export logGLib
func logGLib(cdomain *C.char, logLevel C.GLogLevelFlags, cstr *C.char) {
	str := C.GoString(cstr)

	var domain string
	if cdomain != nil {
		domain = C.GoString(cdomain)
	}
	if len(domain) == 0 {
		domain = "GLib"
	}

	entry := log.WithField("source", domain)

	switch logLevel {
	case C.G_LOG_LEVEL_WARNING:
		entry.Warn(str)
	default:
		entry.Error(str)
	}
}

func Init() {
	C.glib_log_configure()
}
