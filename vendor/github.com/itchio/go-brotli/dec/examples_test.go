package dec

import (
	"io"
	"os"
)

func ExampleBrotliReader() {
	archiveReader, _ := os.Open("data.bin.bro")

	brotliReader := NewBrotliReader(archiveReader)
	defer brotliReader.Close()

	decompressedWriter, _ := os.OpenFile("data.bin.unbro", os.O_CREATE|os.O_WRONLY, 0644)
	defer decompressedWriter.Close()
	io.Copy(decompressedWriter, brotliReader)
}
