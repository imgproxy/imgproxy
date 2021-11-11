package vips

import (
	"sync"

	"github.com/imgproxy/imgproxy/v3/imagetype"
)

func DisableLoadSupport(it imagetype.Type) {
	typeSupportLoad.Store(it, false)
}

func ResetLoadSupport() {
	typeSupportLoad = sync.Map{}
}

func DisableSaveSupport(it imagetype.Type) {
	typeSupportSave.Store(it, false)
}

func ResetSaveSupport() {
	typeSupportSave = sync.Map{}
}
