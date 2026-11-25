package fs

import "net/http"

func NewWithFileSystem(fsys http.FileSystem) *Storage {
	return &Storage{fs: fsys}
}
