package asyncbuffer

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	halfChunkSize   = ChunkSize / 2
	quaterChunkSize = ChunkSize / 4
)

// erraticReader is a test reader that simulates a slow read and can fail after reading a certain number of bytes
type erraticReader struct {
	reader bytes.Reader
	failAt int64 // if set, will return an error after reading this many bytes
}

// Read reads data from the testReader, simulating a slow read and a potential failure
func (r *erraticReader) Read(p []byte) (n int, err error) {
	cur, _ := r.reader.Seek(0, io.SeekCurrent)
	if r.failAt > 0 && r.failAt < cur+int64(len(p)) {
		return 0, errors.New("simulated read failure")
	}
	return r.reader.Read(p)
}

// blockingReader is a test reader which flushes data in chunks
type blockingReader struct {
	reader    bytes.Reader
	mu        sync.Mutex  // locked reader does not return anything
	unlocking atomic.Bool // if true, will proceed without locking each chunk
}

// newBlockingReader creates a new partialReader in locked state
func newBlockingReader(reader bytes.Reader) *blockingReader {
	r := &blockingReader{
		reader: reader,
	}
	r.mu.Lock()
	return r
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

// Read reads data from the testReader, simulating a slow read and a potential failure
func (r *blockingReader) Read(p []byte) (n int, err error) {
	if !r.unlocking.Load() {
		r.mu.Lock()
	}

	n, err = r.reader.Read(p)
	return n, err
}

// generateSourceData generates a byte slice with 4.5 chunks of data
func generateSourceData(t *testing.T, size int64) ([]byte, *bytes.Reader) {
	// We use small chunks for tests, let's check the ChunkSize just in case
	assert.GreaterOrEqual(t, ChunkSize, 20, "ChunkSize required for tests must be greater than 10 bytes")

	// Create a byte slice with 4 chunks of ChunkSize
	source := make([]byte, size)

	// Fill the source with random data
	_, err := rand.Read(source)
	require.NoError(t, err)
	return source, bytes.NewReader(source)
}

// TestAsyncBufferRead tests reading from AsyncBuffer using readAt method which is base for all other methods
func TestAsyncBufferReadAt(t *testing.T) {
	// Let's use source buffer which is 4.5 chunks long
	source, bytesReader := generateSourceData(t, int64(ChunkSize*4)+halfChunkSize)
	asyncBuffer := FromReader(bytesReader)
	defer asyncBuffer.Close()

	asyncBuffer.Wait() // Wait for all chunks to be read since we're going to read all data

	// Let's read all the data
	target := make([]byte, len(source))

	n, err := asyncBuffer.readAt(target, 0)
	require.NoError(t, err)
	assert.Equal(t, len(source), n)
	assert.Equal(t, target, source)

	// Let's read all the data + a bit more
	target = make([]byte, len(source)+1)

	n, err = asyncBuffer.readAt(target, 0)
	require.ErrorIs(t, err, io.EOF) // We read all the data, and reached end
	assert.Equal(t, len(source), n)
	assert.Equal(t, target[:n], source)

	// Let's read > 1 chunk, but with offset from the beginning and the end
	target = make([]byte, len(source)-halfChunkSize)
	n, err = asyncBuffer.readAt(target, quaterChunkSize)
	require.NoError(t, err)
	assert.Equal(t, len(target), n)
	assert.Equal(t, target, source[quaterChunkSize:len(source)-quaterChunkSize])

	// Let's read some data from the middle of the stream < chunk size
	target = make([]byte, ChunkSize/4)
	n, err = asyncBuffer.readAt(target, ChunkSize+ChunkSize/4)
	require.NoError(t, err)
	assert.Equal(t, quaterChunkSize, n)
	assert.Equal(t, target, source[ChunkSize+quaterChunkSize:ChunkSize+quaterChunkSize*2])

	// Let's read some data from the latest half chunk
	target = make([]byte, quaterChunkSize)
	n, err = asyncBuffer.readAt(target, ChunkSize*4+quaterChunkSize)
	require.NoError(t, err)
	assert.Equal(t, quaterChunkSize, n)
	assert.Equal(t, target, source[ChunkSize*4+quaterChunkSize:ChunkSize*4+halfChunkSize])

	// Let's try to read more data then available in the stream
	target = make([]byte, ChunkSize*2)
	n, err = asyncBuffer.readAt(target, ChunkSize*4)
	require.Error(t, err)
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, ChunkSize/2, n)
	assert.Equal(t, target[:ChunkSize/2], source[ChunkSize*4:]) // We read only last half chunk

	// Let's try to read data beyond the end of the stream
	target = make([]byte, ChunkSize*2)
	n, err = asyncBuffer.readAt(target, ChunkSize*5)
	require.Error(t, err)
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, 0, n)
}

// TestAsyncBufferRead tests reading from AsyncBuffer using ReadAt method
func TestAsyncBufferReadAtSmallBuffer(t *testing.T) {
	source, bytesReader := generateSourceData(t, 20)
	asyncBuffer := FromReader(bytesReader)
	defer asyncBuffer.Close()

	// First, let's read all the data
	target := make([]byte, len(source))

	n, err := asyncBuffer.readAt(target, 0)
	require.NoError(t, err)
	assert.Equal(t, len(source), n)
	assert.Equal(t, target, source)

	// Let's read some data
	target = make([]byte, 2)
	n, err = asyncBuffer.readAt(target, 1)
	require.NoError(t, err)
	assert.Equal(t, len(target), n)
	assert.Equal(t, target, source[1:3])

	// Let's read some data beyond the end of the stream
	target = make([]byte, 2)
	n, err = asyncBuffer.readAt(target, 50)
	require.Error(t, err)
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, 0, n)
}

func TestAsyncBufferReader(t *testing.T) {
	source, bytesReader := generateSourceData(t, int64(ChunkSize*4)+halfChunkSize)

	// Create an AsyncBuffer with the byte slice
	asyncBuffer := FromReader(bytesReader)
	defer asyncBuffer.Close()

	// Let's wait for all chunks to be read
	size, err := asyncBuffer.Wait()
	require.NoError(t, err, "AsyncBuffer failed to wait for all chunks")
	assert.Equal(t, int64(ChunkSize*4+halfChunkSize), size)

	reader := asyncBuffer.Reader()

	// Ensure the total length of the data is ChunkSize*4
	require.NoError(t, err)

	// Read the first two chunks
	twoChunks := make([]byte, ChunkSize*2)
	n, err := reader.Read(twoChunks)
	require.NoError(t, err)
	assert.Equal(t, ChunkSize*2, n)
	assert.Equal(t, source[:ChunkSize*2], twoChunks)

	// Seek to the last chunk + 10 bytes
	pos, err := reader.Seek(ChunkSize*3+5, io.SeekStart)
	require.NoError(t, err)
	assert.Equal(t, int64(ChunkSize*3+5), pos)

	// Read the next 10 bytes
	smallSlice := make([]byte, 10)
	n, err = reader.Read(smallSlice)
	require.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.Equal(t, source[ChunkSize*3+5:ChunkSize*3+5+10], smallSlice)

	// Seek -10 bytes from the current position
	pos, err = reader.Seek(-10, io.SeekCurrent)
	require.NoError(t, err)
	assert.Equal(t, int64(ChunkSize*3+5), pos)

	// Read data again
	n, err = reader.Read(smallSlice)
	require.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.Equal(t, source[ChunkSize*3+5:ChunkSize*3+5+10], smallSlice)

	// Seek -10 bytes from end of the stream
	pos, err = reader.Seek(-10, io.SeekEnd)
	require.NoError(t, err)
	assert.Equal(t, size-10, pos)

	// Read last 10 bytes
	n, err = reader.Read(smallSlice)
	require.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.Equal(t, source[size-10:], smallSlice)

	// Seek beyond the end of the stream and try to read
	pos, err = reader.Seek(1024, io.SeekCurrent)
	require.NoError(t, err)
	assert.Equal(t, size+1024, pos)

	_, err = reader.Read(smallSlice)
	require.ErrorIs(t, err, io.EOF)
}

// TestAsyncBufferClose tests closing the AsyncBuffer
func TestAsyncBufferClose(t *testing.T) {
	_, bytesReader := generateSourceData(t, int64(ChunkSize*4)+halfChunkSize)

	// Create an AsyncBuffer with the byte slice
	asyncBuffer := FromReader(bytesReader)

	reader1 := asyncBuffer.Reader()
	reader2 := asyncBuffer.Reader()

	asyncBuffer.Close()

	b := make([]byte, 10)
	_, err := reader1.Read(b)
	require.Error(t, err, "asyncbuffer.AsyncBuffer.ReadAt: attempt to read on closed reader")

	_, err = reader2.Read(b)
	require.Error(t, err, "asyncbuffer.AsyncBuffer.ReadAt: attempt to read on closed reader")

	// After closing the closed reader, it should not panic
	asyncBuffer.Close()

	_, err = reader2.Read(b)
	require.Error(t, err, "asyncbuffer.AsyncBuffer.ReadAt: attempt to read on closed reader")
}

// TestAsyncBufferReadAtErrAtSomePoint tests reading from AsyncBuffer using readAt method
// which would fail somewhere
func TestAsyncBufferReadAtErrAtSomePoint(t *testing.T) {
	// Let's use source buffer which is 4.5 chunks long
	source, bytesReader := generateSourceData(t, int64(ChunkSize*4)+halfChunkSize)
	slowReader := &erraticReader{reader: *bytesReader, failAt: ChunkSize*3 + 5} // fails at last chunk
	asyncBuffer := FromReader(slowReader)
	defer asyncBuffer.Close()

	// Let's wait for all chunks to be read
	_, err := asyncBuffer.Wait()
	require.Error(t, err, "simulated read failure")

	// Let's read something, but before error occurs
	target := make([]byte, halfChunkSize)
	n, err := asyncBuffer.readAt(target, 0)
	require.NoError(t, err)
	assert.Equal(t, len(target), n)
	assert.Equal(t, target, source[:halfChunkSize])

	// And again
	target = make([]byte, halfChunkSize)
	n, err = asyncBuffer.readAt(target, halfChunkSize)
	require.NoError(t, err)
	assert.Equal(t, len(target), n)
	assert.Equal(t, target, source[halfChunkSize:halfChunkSize*2])

	// Let's read something, but when error occurs
	target = make([]byte, halfChunkSize)
	_, err = asyncBuffer.readAt(target, ChunkSize*3)
	require.Error(t, err, "simulated read failure")
}

// TestAsyncBufferReadAsync tests reading from AsyncBuffer using readAt method
// with full = false
func TestAsyncBufferReadAsync(t *testing.T) {
	// Let's use source buffer which is 4.5 chunks long
	source, bytesReader := generateSourceData(t, int64(ChunkSize)*3)
	blockingReader := newBlockingReader(*bytesReader)
	asyncBuffer := FromReader(blockingReader)
	defer asyncBuffer.Close()

	// flush the first chunk to allow reading
	blockingReader.flushNextChunk()

	// Let's try to read first two chunks, however,
	// we know that only the first chunk is available
	target := make([]byte, ChunkSize*2)
	n, err := asyncBuffer.readAt(target, 0)
	require.NoError(t, err)
	assert.Equal(t, ChunkSize, n)
	assert.Equal(t, target[:ChunkSize], source[:ChunkSize])

	blockingReader.flushNextChunk()    // unlock reader to allow read second chunk
	asyncBuffer.WaitFor(ChunkSize + 1) // wait for the second chunk to be available

	target = make([]byte, ChunkSize*2)
	n, err = asyncBuffer.readAt(target, 0)
	require.NoError(t, err)
	assert.Equal(t, ChunkSize*2, n)
	assert.Equal(t, target, source[:ChunkSize*2])

	blockingReader.flush() // Flush the rest of the data
	asyncBuffer.Wait()

	// Try to read near end of the stream, EOF
	target = make([]byte, ChunkSize)
	n, err = asyncBuffer.readAt(target, ChunkSize*3-1)
	require.ErrorIs(t, err, io.EOF)
	assert.Equal(t, 1, n)
	assert.Equal(t, target[0], source[ChunkSize*3-1])

	// Try to read beyond the end of the stream == eof
	target = make([]byte, ChunkSize)
	n, err = asyncBuffer.readAt(target, ChunkSize*3)
	require.ErrorIs(t, io.EOF, err)
	assert.Equal(t, 0, n)
}
