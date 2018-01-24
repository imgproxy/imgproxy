package enc

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

const (
	testQuality int = 9
)

func TestBufferSizes(T *testing.T) {
	options := &BrotliWriterOptions{
		Quality: testQuality,
	}

	input1 := []byte(strings.Repeat("The quick brown fox jumps over the lazy dog", 100000))
	log.Printf("q=%d, inputSize=%d\n", options.Quality, len(input1))

	_, err := CompressBuffer(input1, options)
	if err != nil {
		T.Error(err)
	}
}

func TestStreamEncode(T *testing.T) {
	options := &BrotliWriterOptions{
		Quality: testQuality,
	}

	input1 := []byte(strings.Repeat("The quick brown fox jumps over the lazy dog", 100000))
	inputSize := len(input1)
	log.Printf("q=%d, inputSize=%d\n", options.Quality, inputSize)

	for lgwin := 16; lgwin <= 22; lgwin++ {
		options.LGWin = lgwin

		// compress the entire data in one go
		fullBufferOutput, err := CompressBuffer(input1, options)
		if err != nil {
			T.Error(err)
		}

		// then using the high-level Writer interface
		writerBuffer := new(bytes.Buffer)
		writer := NewBrotliWriter(writerBuffer, options)
		writer.Write(input1)
		writer.Close()

		fullWriterOutput := writerBuffer.Bytes()
		if !bytes.Equal(fullWriterOutput, fullBufferOutput) {
			T.Fatalf("for lgwin %d, stream writer compression didn't give same result as buffer compression", options.LGWin)
		}

		outputSize := len(fullWriterOutput)
		log.Printf("lgwin=%d, output=%d (%.4f%% of input size)\n", options.LGWin, outputSize, float32(outputSize)*100.0/float32(inputSize))
	}
}
