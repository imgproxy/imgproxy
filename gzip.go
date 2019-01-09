package main

import (
	"compress/gzip"
	"io"
	"os"
	"sync"
)

var nullwriter, _ = os.Open("/dev/null")

var gzipPool = sync.Pool{
	New: func() interface{} {
		gz, _ := gzip.NewWriterLevel(nullwriter, conf.GZipCompression)
		return gz
	},
}

func gzipData(data []byte, w io.Writer) {
	gz := gzipPool.Get().(*gzip.Writer)
	defer gzipPool.Put(gz)

	gz.Reset(w)
	gz.Write(data)
	gz.Close()
}
