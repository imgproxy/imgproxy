package vips

import "github.com/imgproxy/imgproxy/v2/imagetype"

func DisableLoadSupport(it imagetype.Type) {
	typeSupportLoad[it] = false
}

func ResetLoadSupport() {
	typeSupportLoad = make(map[imagetype.Type]bool)
}

func DisableSaveSupport(it imagetype.Type) {
	typeSupportSave[it] = false
}

func ResetSaveSupport() {
	typeSupportSave = make(map[imagetype.Type]bool)
}
