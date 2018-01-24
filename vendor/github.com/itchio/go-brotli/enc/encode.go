// Package enc provides Brotli encoder bindings
package enc // import "github.com/itchio/go-brotli/enc"

/*
#cgo CFLAGS: -I${SRCDIR}/../include
#cgo LDFLAGS: -lm

// for memcpy
#include <string.h>

#include <brotli/encode.h>

void* encodeBrotliDictionary;

struct CompressStreamResult {
  size_t bytes_consumed;
  const uint8_t* output_data;
  size_t output_data_size;
  int success;
  int has_more;
};
static struct CompressStreamResult CompressStream(
    BrotliEncoderState* s, BrotliEncoderOperation op,
    const uint8_t* data, size_t data_size) {
  struct CompressStreamResult result;
  size_t available_in = data_size;
  const uint8_t* next_in = data;
  size_t available_out = 0;
  result.success = BrotliEncoderCompressStream(s, op,
      &available_in, &next_in, &available_out, 0, 0) ? 1 : 0;
  result.bytes_consumed = data_size - available_in;
  result.output_data = 0;
  result.output_data_size = 0;
  if (result.success) {
    result.output_data = BrotliEncoderTakeOutput(s, &result.output_data_size);
  }
  result.has_more = BrotliEncoderHasMoreOutput(s) ? 1 : 0;
  return result;
}
*/
import "C"

import (
	"bytes"
	"errors"
	"io"
	"runtime"
	"unsafe"

	"github.com/itchio/go-brotli/common"
)

// Errors which may be returned when encoding
var (
	errWriterClosed = errors.New("brotli-go: Writer is closed")
	errEncode       = errors.New("brotli-go: encode error")
)

func init() {
	// Set up the default dictionary from the data in the shared package
	C.encodeBrotliDictionary = unsafe.Pointer(common.GetDictionary())
}

// BrotliWriterOptions configures BrotliWriter
type BrotliWriterOptions struct {
	// Quality controls the compression-speed vs compression-density trade-offs.
	// The higher the quality, the slower the compression. Range is 0 to 11.
	Quality int

	// LGWin is the base 2 logarithm of the sliding window size.
	// Range is 10 to 24. 0 indicates automatic configuration based on Quality.
	LGWin int
}

// CompressBuffer compresses a single block of data.
// Default options are used if options is nil.
// Returns the brotli-compressed version of the data, or an error.
func CompressBuffer(input []byte, options *BrotliWriterOptions) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := NewBrotliWriter(buf, options)
	_, err := w.Write(input)
	if err != nil {
		return nil, err
	}

	err = w.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// BrotliWriter implements the io.Writer interface, compressing the stream
// to an output Writer using Brotli.
type BrotliWriter struct {
	dst io.Writer

	state        *C.BrotliEncoderState
	buf, encoded []byte
}

// NewBrotliWriter instantiates a new BrotliWriter that writes with dst,
// and configured with the passed options (which may be nil)
func NewBrotliWriter(dst io.Writer, options *BrotliWriterOptions) *BrotliWriter {
	state := C.BrotliEncoderCreateInstance(nil, nil, nil)
	w := &BrotliWriter{
		dst:   dst,
		state: state,
	}
	runtime.SetFinalizer(w, brotliWriterFinalizer)

	if options != nil {
		C.BrotliEncoderSetParameter(state, C.BROTLI_PARAM_QUALITY, C.uint32_t(options.Quality))

		if options.LGWin > 0 {
			C.BrotliEncoderSetParameter(state, C.BROTLI_PARAM_LGWIN, C.uint32_t(options.LGWin))
		}
	}

	return w
}

func brotliWriterFinalizer(bw *BrotliWriter) {
	// not a terribly good idea (swallows errors),
	// but at least avoids memory leaks?
	bw.Close()
}

func (w *BrotliWriter) Write(buffer []byte) (int, error) {
	return w.writeChunk(buffer, C.BROTLI_OPERATION_PROCESS)
}

// Close cleans up the resources used by the Brotli encoder for this
// stream. If the output buffer is an io.Closer, it will also be closed.
func (w *BrotliWriter) Close() error {
	// If stream is already closed, it is reported by `writeChunk`.
	_, err := w.writeChunk(nil, C.BROTLI_OPERATION_FINISH)
	// C-Brotli tolerates `nil` pointer here.
	C.BrotliEncoderDestroyInstance(w.state)
	w.state = nil
	if err != nil {
		return err
	}

	if v, ok := w.dst.(io.Closer); ok {
		return v.Close()
	}

	return nil
}

func (w *BrotliWriter) writeChunk(p []byte, op C.BrotliEncoderOperation) (n int, err error) {
	if w.state == nil {
		return 0, errWriterClosed
	}

	for {
		var data *C.uint8_t
		if len(p) != 0 {
			data = (*C.uint8_t)(&p[0])
		}
		result := C.CompressStream(w.state, op, data, C.size_t(len(p)))
		if result.success == 0 {
			return n, errEncode
		}
		p = p[int(result.bytes_consumed):]
		n += int(result.bytes_consumed)

		length := int(result.output_data_size)
		if length != 0 {
			// It is a workaround for non-copying-wrapping of native memory.
			// C-encoder never pushes output block longer than ((2 << 25) + 502).
			// TODO: use natural wrapper, when it becomes available, see
			//               https://golang.org/issue/13656.
			output := (*[1 << 30]byte)(unsafe.Pointer(result.output_data))[:length:length]
			_, err = w.dst.Write(output)
			if err != nil {
				return n, err
			}
		}
		if len(p) == 0 && result.has_more == 0 {
			return n, nil
		}
	}
}

// internal cgo utilities

func toC(array []byte) *C.uint8_t {
	return (*C.uint8_t)(unsafe.Pointer(&array[0]))
}
