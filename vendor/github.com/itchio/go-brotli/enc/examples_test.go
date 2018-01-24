package enc

import (
	"io"
	"os"
)

func ExampleBrotliWriterOptions() {
	var input []byte

	compressed, _ := CompressBuffer(input, &BrotliWriterOptions{
		// brotli supports quality values from 0 to 11 included
		// 0 is the fastest, 11 is the most compressed but slowest
		Quality: 0,
	})
	_ = compressed
}

func ExampleBrotliWriter() {
	compressedWriter, _ := os.OpenFile("data.bin.bro", os.O_CREATE|os.O_WRONLY, 0644)

	brotliWriter := NewBrotliWriter(compressedWriter, nil)
	// BrotliWriter will close writer passed as argument if it implements io.Closer
	defer brotliWriter.Close()

	fileReader, _ := os.Open("data.bin")
	defer fileReader.Close()

	io.Copy(brotliWriter, fileReader)
}
