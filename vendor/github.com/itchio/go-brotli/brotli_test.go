package brotli

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	"github.com/itchio/go-brotli/dec"
	"github.com/itchio/go-brotli/enc"
)

func TestSimpleString(T *testing.T) {
	testCompress([]byte("Hello Hello Hello, Hello Hello Hello"), T)
}

func TestShortString(T *testing.T) {
	s := []byte("The quick brown fox")
	l := len(s)

	// Brotli will not compress arrays shorter than 3 characters
	for ; l > 3; l-- {
		testCompress(s[:l], T)
	}
}

func testCompress(s []byte, T *testing.T) {
	T.Logf("Compressing: %s\n", s)

	encoded, cerr := enc.CompressBuffer(s, nil)
	if cerr != nil {
		T.Error(cerr)
	}

	decoded, derr := dec.DecompressBuffer(encoded, nil)
	if derr != nil {
		T.Error(derr)
	}

	if !bytes.Equal(s, decoded) {
		T.Logf("Decompressed: %s\n", decoded)
		T.Error("Decoded output does not match original input")
	}
}

// Run roundtrip tests from Brotli repository
func TestRoundtrip(T *testing.T) {
	inputs := []string{
		"testdata/alice29.txt",
		"testdata/asyoulik.txt",
		"testdata/lcet10.txt",
		"testdata/plrabn12.txt",
		"enc/encode.c",
		"common/dictionary.h",
		"dec/decode.c",
	}

	for _, file := range inputs {
		var err error
		var input []byte

		input, err = ioutil.ReadFile(file)
		if err != nil {
			T.Error(err)
		}

		for _, quality := range []int{1, 6, 9, 11} {
			T.Logf("Roundtrip testing %s at quality %d", file, quality)

			options := &enc.BrotliWriterOptions{
				Quality: quality,
			}

			bro := testCompressBuffer(options, input, T)

			testDecompressBuffer(input, bro, T)

			testDecompressStream(input, bytes.NewReader(bro), T)

			// Stream compress
			buffer := new(bytes.Buffer)
			testCompressStream(options, input, buffer, T)

			testDecompressBuffer(input, buffer.Bytes(), T)

			// Stream roundtrip
			reader, writer := io.Pipe()
			go testCompressStream(options, input, writer, T)
			testDecompressStream(input, reader, T)
		}
	}
}

func testCompressBuffer(options *enc.BrotliWriterOptions, input []byte, T *testing.T) []byte {
	// Test buffer compression
	bro, err := enc.CompressBuffer(input, options)
	if err != nil {
		T.Error(err)
	}
	T.Logf("  Compressed from %d to %d bytes, %.1f%%", len(input), len(bro), (float32(len(bro))/float32(len(input)))*100)

	return bro
}

func testDecompressBuffer(input, bro []byte, T *testing.T) {
	// Buffer decompression
	unbro, err := dec.DecompressBuffer(bro, nil)
	if err != nil {
		T.Error(err)
	}

	check("Buffer decompress", input, unbro, T)
}

func testDecompressStream(input []byte, reader io.Reader, T *testing.T) {
	// Stream decompression - use ridiculously small buffer on purpose to
	// test NEEDS_MORE_INPUT state, cf. https://github.com/kothar/brotli-go/issues/28
	streamUnbro, err := ioutil.ReadAll(dec.NewBrotliReaderSize(reader, 128))
	if err != nil {
		T.Error(err)
	}

	check("Stream decompress", input, streamUnbro, T)
}

func testCompressStream(options *enc.BrotliWriterOptions, input []byte, writer io.Writer, T *testing.T) {
	bwriter := enc.NewBrotliWriter(writer, options)
	n, err := bwriter.Write(input)
	if err != nil {
		T.Error(err)
	}

	err = bwriter.Close()
	if err != nil {
		T.Error(err)
	}

	if n != len(input) {
		T.Error("Not all input was consumed")
	}
}

func check(test string, input, output []byte, T *testing.T) {
	if len(input) != len(output) {
		T.Errorf("  %s: Length of decompressed output (%d) doesn't match input (%d)", test, len(output), len(input))
	}

	if !bytes.Equal(input, output) {
		T.Errorf("  %s: Input does not match decompressed output", test)
	}
}
