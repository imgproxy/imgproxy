package asyncbuffer_test

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/imgproxy/imgproxy/v4/asyncbuffer"
)

const (
	halfChunkSize   = asyncbuffer.ChunkSize / 2
	quaterChunkSize = asyncbuffer.ChunkSize / 4
)

// countingReader wraps an io.ReadSeeker and counts the total bytes read from it
type countingReader struct {
	r io.Reader
	n atomic.Int64
}

func newCountingReader(r io.Reader) *countingReader {
	return &countingReader{r: r}
}

func (r *countingReader) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	r.n.Add(int64(n))
	return n, err
}

func (r *countingReader) Close() error { return nil }

func (r *countingReader) BytesRead() int64 {
	return r.n.Load()
}

// erraticReader is a test reader that simulates a slow read and can fail after reading a certain number of bytes
type erraticReader struct {
	reader *countingReader
	failAt int64 // if set, will return an error after reading this many bytes
}

// Read reads data from the testReader, simulating a slow read and a potential failure
func (r *erraticReader) Read(p []byte) (n int, err error) {
	cur := r.reader.BytesRead()
	if r.failAt > 0 && r.failAt < cur+int64(len(p)) {
		return 0, errors.New("simulated read failure")
	}
	return r.reader.Read(p)
}

// Close forwards closing to the underlying reader
func (r *erraticReader) Close() error {
	return r.reader.Close()
}

// blockingReader is a test reader which flushes data in chunks
type blockingReader struct {
	reader    io.ReadCloser
	mu        sync.Mutex  // locked reader does not return anything
	unlocking atomic.Bool // if true, will proceed without locking each chunk
}

// newBlockingReader creates a new partialReader in locked state
func newBlockingReader(reader io.ReadCloser) *blockingReader {
	r := &blockingReader{
		reader: reader,
	}
	r.mu.Lock()
	return r
}

// Read reads data from the testReader, simulating a slow read and a potential failure
func (r *blockingReader) Read(p []byte) (n int, err error) {
	if !r.unlocking.Load() {
		r.mu.Lock()
	}

	n, err = r.reader.Read(p)
	return n, err
}

func (r *blockingReader) Close() error { // Close forwards closing to the underlying reader
	return r.reader.Close()
}

// flushNextChunk unlocks the reader, allowing it to return the next chunk of data
func (r *blockingReader) flushNextChunk() {
	r.mu.Unlock()
}

// flush unlocks the reader, allowing it to return all data as usual
func (r *blockingReader) flush() {
	r.unlocking.Store(true) // allow reading data without blocking
	r.mu.Unlock()           // and continue
}

// generateSourceData generates a byte slice with 4.5 chunks of data
func generateSourceData(t *testing.T, size int) ([]byte, *countingReader) {
	t.Helper()

	// We use small chunks for tests, let's check the ChunkSize just in case
	assert.GreaterOrEqual(t, asyncbuffer.ChunkSize, 20, "ChunkSize required for tests must be greater than 10 bytes")

	// Create a byte slice with 4 chunks of ChunkSize
	source := make([]byte, size)

	// Fill the source with random data
	_, err := rand.Read(source)
	require.NoError(t, err)
	return source, newCountingReader(bytes.NewReader(source))
}

// TestAsyncBufferRead tests reading from AsyncBuffer using readAt method which is base for all other methods
func TestAsyncBufferReadAt(t *testing.T) {
	// Let's use source buffer which is 4.5 chunks long
	source, bytesReader := generateSourceData(t, asyncbuffer.ChunkSize*4+halfChunkSize)
	ab := asyncbuffer.New(bytesReader, -1)
	defer ab.Close()

	ab.Wait() // Wait for all chunks to be read since we're going to read all data

	// Let's read all the data
	target := make([]byte, len(source))

	n, err := ab.ReadAt(target, 0)
	require.NoError(t, err)
	assert.Equal(t, len(source), n)
	assert.Equal(t, target, source)

	// Let's read all the data + a bit more
	target = make([]byte, len(source)+1)

	n, err = ab.ReadAt(target, 0)
	require.NoError(t, err) // We read all the data, and reached end
	assert.Equal(t, len(source), n)
	assert.Equal(t, target[:n], source)

	// Let's read > 1 chunk, but with offset from the beginning and the end
	target = make([]byte, len(source)-halfChunkSize)
	n, err = ab.ReadAt(target, quaterChunkSize)
	require.NoError(t, err)
	assert.Equal(t, len(target), n)
	assert.Equal(t, target, source[quaterChunkSize:len(source)-quaterChunkSize])

	// Let's read some data from the middle of the stream < chunk size
	target = make([]byte, asyncbuffer.ChunkSize/4)
	n, err = ab.ReadAt(target, asyncbuffer.ChunkSize+asyncbuffer.ChunkSize/4)
	require.NoError(t, err)
	assert.Equal(t, quaterChunkSize, n)
	assert.Equal(t, target, source[asyncbuffer.ChunkSize+quaterChunkSize:asyncbuffer.ChunkSize+quaterChunkSize*2])

	// Let's read some data from the latest half chunk
	target = make([]byte, quaterChunkSize)
	n, err = ab.ReadAt(target, asyncbuffer.ChunkSize*4+quaterChunkSize)
	require.NoError(t, err)
	assert.Equal(t, quaterChunkSize, n)
	assert.Equal(t, target, source[asyncbuffer.ChunkSize*4+quaterChunkSize:asyncbuffer.ChunkSize*4+halfChunkSize])

	// Let's try to read more data then available in the stream
	target = make([]byte, asyncbuffer.ChunkSize*2)
	n, err = ab.ReadAt(target, asyncbuffer.ChunkSize*4)
	require.NoError(t, err)
	assert.Equal(t, asyncbuffer.ChunkSize/2, n)
	assert.Equal(t, target[:asyncbuffer.ChunkSize/2], source[asyncbuffer.ChunkSize*4:]) // We read only last half chunk

	// Let's try to read data beyond the end of the stream
	target = make([]byte, asyncbuffer.ChunkSize*2)
	n, err = ab.ReadAt(target, asyncbuffer.ChunkSize*5)
	require.Error(t, err)
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, 0, n)
}

// TestAsyncBufferRead tests reading from AsyncBuffer using ReadAt method
func TestAsyncBufferReadAtSmallBuffer(t *testing.T) {
	source, bytesReader := generateSourceData(t, 20)
	ab := asyncbuffer.New(bytesReader, -1)
	defer ab.Close()

	// First, let's read all the data
	target := make([]byte, len(source))

	n, err := ab.ReadAt(target, 0)
	require.NoError(t, err)
	assert.Equal(t, len(source), n)
	assert.Equal(t, target, source)

	// Let's read some data
	target = make([]byte, 2)
	n, err = ab.ReadAt(target, 1)
	require.NoError(t, err)
	assert.Equal(t, len(target), n)
	assert.Equal(t, target, source[1:3])

	// Let's read some data beyond the end of the stream
	target = make([]byte, 2)
	n, err = ab.ReadAt(target, 50)
	require.Error(t, err)
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, 0, n)
}

func TestAsyncBufferReader(t *testing.T) {
	source, bytesReader := generateSourceData(t, asyncbuffer.ChunkSize*4+halfChunkSize)

	// Create an AsyncBuffer with the byte slice
	ab := asyncbuffer.New(bytesReader, -1)
	defer ab.Close()

	// Let's wait for all chunks to be read
	size, err := ab.Wait()
	require.NoError(t, err, "AsyncBuffer failed to wait for all chunks")
	assert.Equal(t, asyncbuffer.ChunkSize*4+halfChunkSize, size)

	reader := ab.Reader()

	// Ensure the total length of the data is ChunkSize*4
	require.NoError(t, err)

	// Read the first two chunks
	twoChunks := make([]byte, asyncbuffer.ChunkSize*2)
	n, err := reader.Read(twoChunks)
	require.NoError(t, err)
	assert.Equal(t, asyncbuffer.ChunkSize*2, n)
	assert.Equal(t, source[:asyncbuffer.ChunkSize*2], twoChunks)

	// Seek to the last chunk + 10 bytes
	pos, err := reader.Seek(asyncbuffer.ChunkSize*3+5, io.SeekStart)
	require.NoError(t, err)
	assert.Equal(t, int64(asyncbuffer.ChunkSize*3+5), pos)

	// Read the next 10 bytes
	smallSlice := make([]byte, 10)
	n, err = reader.Read(smallSlice)
	require.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.Equal(t, source[asyncbuffer.ChunkSize*3+5:asyncbuffer.ChunkSize*3+5+10], smallSlice)

	// Seek -10 bytes from the current position
	pos, err = reader.Seek(-10, io.SeekCurrent)
	require.NoError(t, err)
	assert.Equal(t, int64(asyncbuffer.ChunkSize*3+5), pos)

	// Read data again
	n, err = reader.Read(smallSlice)
	require.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.Equal(t, source[asyncbuffer.ChunkSize*3+5:asyncbuffer.ChunkSize*3+5+10], smallSlice)

	// Seek -10 bytes from end of the stream
	pos, err = reader.Seek(-10, io.SeekEnd)
	require.NoError(t, err)
	assert.Equal(t, size-10, int(pos))

	// Read last 10 bytes
	n, err = reader.Read(smallSlice)
	require.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.Equal(t, source[size-10:], smallSlice)

	// Seek beyond the end of the stream and try to read
	pos, err = reader.Seek(1024, io.SeekCurrent)
	require.NoError(t, err)
	assert.Equal(t, size+1024, int(pos))

	_, err = reader.Read(smallSlice)
	require.ErrorIs(t, err, io.EOF)
}

// TestAsyncBufferClose tests closing the AsyncBuffer
func TestAsyncBufferClose(t *testing.T) {
	_, bytesReader := generateSourceData(t, asyncbuffer.ChunkSize*4+halfChunkSize)

	// Create an AsyncBuffer with the byte slice
	ab := asyncbuffer.New(bytesReader, -1)

	reader1 := ab.Reader()
	reader2 := ab.Reader()

	ab.Close()

	b := make([]byte, 10)
	_, err := reader1.Read(b)
	require.Error(t, err, "asyncbuffer.AsyncBuffer.ReadAt: attempt to read on closed reader")

	_, err = reader2.Read(b)
	require.Error(t, err, "asyncbuffer.AsyncBuffer.ReadAt: attempt to read on closed reader")

	// After closing the closed reader, it should not panic
	ab.Close()

	_, err = reader2.Read(b)
	require.Error(t, err, "asyncbuffer.AsyncBuffer.ReadAt: attempt to read on closed reader")
}

// TestAsyncBufferReadAtErrAtSomePoint tests reading from AsyncBuffer using readAt method
// which would fail somewhere
func TestAsyncBufferReadAtErrAtSomePoint(t *testing.T) {
	// Let's use source buffer which is 4.5 chunks long
	source, bytesReader := generateSourceData(t, asyncbuffer.ChunkSize*4+halfChunkSize)
	slowReader := &erraticReader{reader: bytesReader, failAt: asyncbuffer.ChunkSize*3 + 5} // fails at last chunk
	ab := asyncbuffer.New(slowReader, -1)
	defer ab.Close()

	// Let's wait for all chunks to be read
	_, err := ab.Wait()
	require.Error(t, err, "simulated read failure")

	// Let's read something, but before error occurs
	target := make([]byte, halfChunkSize)
	n, err := ab.ReadAt(target, 0)
	require.NoError(t, err)
	assert.Equal(t, len(target), n)
	assert.Equal(t, target, source[:halfChunkSize])

	// And again
	target = make([]byte, halfChunkSize)
	n, err = ab.ReadAt(target, halfChunkSize)
	require.NoError(t, err)
	assert.Equal(t, len(target), n)
	assert.Equal(t, target, source[halfChunkSize:halfChunkSize*2])

	// Let's read something, but when error occurs
	target = make([]byte, halfChunkSize)
	_, err = ab.ReadAt(target, asyncbuffer.ChunkSize*3)
	require.Error(t, err, "simulated read failure")
}

// TestAsyncBufferReadAsync tests reading from AsyncBuffer using readAt method
// with full = false
func TestAsyncBufferReadAsync(t *testing.T) {
	// Let's use source buffer which is 4.5 chunks long
	source, bytesReader := generateSourceData(t, asyncbuffer.ChunkSize*3)
	br := newBlockingReader(bytesReader)
	ab := asyncbuffer.New(br, -1)
	defer ab.Close()

	// flush the first chunk to allow reading
	br.flushNextChunk()

	// Let's try to read first two chunks, however,
	// we know that only the first chunk is available
	target := make([]byte, asyncbuffer.ChunkSize*2)
	n, err := ab.ReadAt(target, 0)
	require.NoError(t, err)
	assert.Equal(t, asyncbuffer.ChunkSize, n)
	assert.Equal(t, target[:asyncbuffer.ChunkSize], source[:asyncbuffer.ChunkSize])

	br.flushNextChunk()                   // unlock reader to allow read second chunk
	ab.WaitFor(asyncbuffer.ChunkSize + 1) // wait for the second chunk to be available

	target = make([]byte, asyncbuffer.ChunkSize*2)
	n, err = ab.ReadAt(target, 0)
	require.NoError(t, err)
	assert.Equal(t, asyncbuffer.ChunkSize*2, n)
	assert.Equal(t, target, source[:asyncbuffer.ChunkSize*2])

	br.flush() // Flush the rest of the data
	ab.Wait()

	// Try to read near end of the stream, EOF
	target = make([]byte, asyncbuffer.ChunkSize)
	n, err = ab.ReadAt(target, asyncbuffer.ChunkSize*3-1)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, target[0], source[asyncbuffer.ChunkSize*3-1])

	// Try to read beyond the end of the stream == eof
	target = make([]byte, asyncbuffer.ChunkSize)
	n, err = ab.ReadAt(target, asyncbuffer.ChunkSize*3)
	require.ErrorIs(t, io.EOF, err)
	assert.Equal(t, 0, n)
}

// TestAsyncBufferWithDataLenAndExactReaderSize tests that AsyncBuffer doesn't
// return an error when the expected data length is set and matches the reader size
func TestAsyncBufferWithDataLenAndExactReaderSize(t *testing.T) {
	source, bytesReader := generateSourceData(t, asyncbuffer.ChunkSize*4+halfChunkSize)
	ab := asyncbuffer.New(bytesReader, len(source))
	defer ab.Close()

	// Let's wait for all chunks to be read
	size, err := ab.Wait()
	require.NoError(t, err, "AsyncBuffer failed to wait for all chunks")
	assert.Equal(t, len(source), size)
}

// TestAsyncBufferWithDataLenAndShortReaderSize tests that AsyncBuffer returns
// io.ErrUnexpectedEOF when the expected data length is set and the reader size
// is shorter than the expected data length
func TestAsyncBufferWithDataLenAndShortReaderSize(t *testing.T) {
	source, bytesReader := generateSourceData(t, asyncbuffer.ChunkSize*4+halfChunkSize)
	ab := asyncbuffer.New(bytesReader, len(source)+100) // 100 bytes more than the source
	defer ab.Close()

	// Let's wait for all chunks to be read
	size, err := ab.Wait()
	require.Equal(t, len(source), size)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF,
		"AsyncBuffer should return io.ErrUnexpectedEOF when data length is longer than reader size")
}

// TestAsyncBufferWithDataLenAndLongerReaderSize tests that AsyncBuffer doesn't
// read more data than specified by the expected data length and doesn't return an error
// when the reader size is longer than the expected data length
func TestAsyncBufferWithDataLenAndLongerReaderSize(t *testing.T) {
	source, bytesReader := generateSourceData(t, asyncbuffer.ChunkSize*4+halfChunkSize)
	ab := asyncbuffer.New(bytesReader, len(source)-100) // 100 bytes less than the source
	defer ab.Close()

	// Let's wait for all chunks to be read
	size, err := ab.Wait()
	require.NoError(t, err, "AsyncBuffer failed to wait for all chunks")
	assert.Equal(t, len(source)-100, size,
		"AsyncBuffer should read only the specified amount of data when data length is set")
}

// TestAsyncBufferReadAllCompability tests that ReadAll methods works as expected
func TestAsyncBufferReadAllCompability(t *testing.T) {
	source, err := os.ReadFile("../testdata/test1.jpg")
	require.NoError(t, err)
	ab := asyncbuffer.New(newCountingReader(bytes.NewReader(source)), -1)
	defer ab.Close()

	b, err := io.ReadAll(ab.Reader())
	require.NoError(t, err)
	require.Len(t, b, len(source))
}

func TestAsyncBufferThreshold(t *testing.T) {
	_, bytesReader := generateSourceData(t, asyncbuffer.PauseThreshold*3)
	ab := asyncbuffer.New(bytesReader, -1)
	defer ab.Close()

	target := make([]byte, asyncbuffer.ChunkSize)
	n, err := ab.ReadAt(target, 0)
	require.NoError(t, err)
	assert.Equal(t, asyncbuffer.ChunkSize, n)

	// Ensure that buffer hits the pause threshold
	require.Eventually(t, func() bool {
		return bytesReader.BytesRead() >= asyncbuffer.PauseThreshold
	}, 300*time.Millisecond, 10*time.Millisecond)

	// Ensure that buffer never reaches the end of the stream
	require.Never(t, func() bool {
		return bytesReader.BytesRead() >= asyncbuffer.PauseThreshold*2-1
	}, 300*time.Millisecond, 10*time.Millisecond)

	// Let's hit the pause threshold
	target = make([]byte, asyncbuffer.PauseThreshold)
	n, err = ab.ReadAt(target, 0)
	require.NoError(t, err)
	require.Equal(t, asyncbuffer.PauseThreshold, n)

	// Ensure that buffer never reaches the end of the stream
	require.Never(t, func() bool {
		return bytesReader.BytesRead() >= asyncbuffer.PauseThreshold*2-1
	}, 300*time.Millisecond, 10*time.Millisecond)

	// Let's hit the pause threshold
	target = make([]byte, asyncbuffer.PauseThreshold+1)
	n, err = ab.ReadAt(target, 0)
	require.NoError(t, err)

	// It usually returns only pauseThreshold bytes because this exact operation unpauses the reader,
	// but the initial offset is before the threshold, data beyond the threshold may not be available.
	assert.GreaterOrEqual(t, asyncbuffer.PauseThreshold, n)

	// Ensure that buffer hits the end of the stream
	require.Eventually(t, func() bool {
		return bytesReader.BytesRead() >= asyncbuffer.PauseThreshold*2
	}, 300*time.Millisecond, 10*time.Millisecond)
}

func TestAsyncBufferThresholdInstantBeyondAccess(t *testing.T) {
	_, bytesReader := generateSourceData(t, asyncbuffer.PauseThreshold*3)
	ab := asyncbuffer.New(bytesReader, -1)
	defer ab.Close()

	target := make([]byte, asyncbuffer.ChunkSize)
	n, err := ab.ReadAt(target, asyncbuffer.PauseThreshold+1)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, asyncbuffer.ChunkSize, n)

	// Ensure that buffer hits the end of the stream
	require.Eventually(t, func() bool {
		return bytesReader.BytesRead() >= asyncbuffer.PauseThreshold*2
	}, 300*time.Millisecond, 10*time.Millisecond)
}
