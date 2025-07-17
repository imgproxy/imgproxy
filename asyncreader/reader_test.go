package asyncreader

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	halfChunkSize   = ChunkSize / 2
	quaterChunkSize = ChunkSize / 4
)

type slowReader struct {
	reader bytes.Reader
}

func (sr *slowReader) Read(p []byte) (n int, err error) {
	time.Sleep(500 * time.Millisecond) // artificially delay reading
	return sr.reader.Read(p)
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

// TestAsyncReaderRead tests reading from AsyncReader using readAt method which is base for all other methods
func TestAsyncReaderReadAt(t *testing.T) {
	// Let's use source buffer which is 4.5 chunks long
	source, bytesReader := generateSourceData(t, int64(ChunkSize*4)+halfChunkSize)
	asyncReader := NewAsyncReader(bytesReader)
	defer asyncReader.Close()

	// Let's read all the data
	target := make([]byte, len(source))

	n, err := asyncReader.readAt(target, 0)
	require.NoError(t, err)
	assert.Equal(t, len(source), n)
	assert.Equal(t, target, source)

	// Let's read > 1 chunk, but with offset from the beginning and the end
	target = make([]byte, len(source)-halfChunkSize)
	n, err = asyncReader.readAt(target, quaterChunkSize)
	require.NoError(t, err)
	assert.Equal(t, len(target), n)
	assert.Equal(t, target, source[quaterChunkSize:len(source)-quaterChunkSize])

	// Let's read some data from the middle of the stream < chunk size
	target = make([]byte, ChunkSize/4)
	n, err = asyncReader.readAt(target, ChunkSize+ChunkSize/4)
	require.NoError(t, err)
	assert.Equal(t, quaterChunkSize, n)
	assert.Equal(t, target, source[ChunkSize+quaterChunkSize:ChunkSize+quaterChunkSize*2])

	// Let's try to read more data then available in the stream
	target = make([]byte, ChunkSize*2)
	n, err = asyncReader.readAt(target, ChunkSize*4)
	require.Error(t, err)
	assert.Equal(t, err, io.ErrUnexpectedEOF)
	assert.Equal(t, ChunkSize, n)
	assert.Equal(t, target[:ChunkSize/2], source[ChunkSize*4:]) // We read only last half chunk

	// Let's try to read data beyond the end of the stream
	target = make([]byte, ChunkSize*2)
	n, err = asyncReader.readAt(target, ChunkSize*5)
	require.Error(t, err)
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, 0, n)
}

// TestAsyncReaderRead tests reading from AsyncReader using ReadAt method
func TestAsyncReaderReadAtSmallBuffer(t *testing.T) {
	source, bytesReader := generateSourceData(t, 20)
	asyncReader := NewAsyncReader(bytesReader)
	defer asyncReader.Close()

	// First, let's read all the data
	target := make([]byte, len(source))

	n, err := asyncReader.readAt(target, 0)
	require.NoError(t, err)
	assert.Equal(t, len(source), n)
	assert.Equal(t, target, source)

	// Let's read some data
	target = make([]byte, 2)
	n, err = asyncReader.readAt(target, 1)
	require.NoError(t, err)
	assert.Equal(t, len(target), n)
	assert.Equal(t, target, source[1:3])

	// Let's read some data beyond the end of the stream
	target = make([]byte, 2)
	n, err = asyncReader.readAt(target, 50)
	require.Error(t, err)
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, 0, n)
}

func TestAsyncReaderReader(t *testing.T) {
	source, bytesReader := generateSourceData(t, int64(ChunkSize*4)+halfChunkSize)

	// Create an AsyncReader with the byte slice
	asyncReader := NewAsyncReader(bytesReader)
	defer asyncReader.Close()

	// Let's wait for all chunks to be read
	err := asyncReader.Wait()
	require.NoError(t, err, "AsyncReader failed to wait for all chunks")

	reader := asyncReader.Reader()

	// Ensure the total length of the data is ChunkSize*4
	size, err := reader.Size()
	require.NoError(t, err)
	assert.Equal(t, int64(ChunkSize*4+halfChunkSize), size)

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
}

// TestAsyncReaderClose tests closing the AsyncReader
func TestAsyncReaderClose(t *testing.T) {
	_, bytesReader := generateSourceData(t, int64(ChunkSize*4)+halfChunkSize)

	// Create an AsyncReader with the byte slice
	asyncReader := NewAsyncReader(bytesReader)

	reader1 := asyncReader.Reader()
	reader2 := asyncReader.Reader()

	asyncReader.Close()

	b := make([]byte, 10)
	_, err := reader1.Read(b)
	require.Error(t, err, "asyncreader.AsyncReader.ReadAt: attempt to read on closed reader")

	_, err = reader2.Read(b)
	require.Error(t, err, "asyncreader.AsyncReader.ReadAt: attempt to read on closed reader")

	// After closing the closed reader, it should not panic
	asyncReader.Close()

	_, err = reader2.Read(b)
	require.Error(t, err, "asyncreader.AsyncReader.ReadAt: attempt to read on closed reader")
}

// TestAsyncReaderRead tests reading from AsyncReader using readAt method which is base for all other methods
func TestAsyncReaderReadAtSlow(t *testing.T) {
	// Let's use source buffer which is 4.5 chunks long
	source, bytesReader := generateSourceData(t, int64(ChunkSize*4)+halfChunkSize)
	slowReader := &slowReader{reader: *bytesReader}
	asyncReader := NewAsyncReader(slowReader)
	defer asyncReader.Close()

	// Let's read > 1 chunk, but with offset from the beginning and the end
	target := make([]byte, len(source)-halfChunkSize)
	n, err := asyncReader.readAt(target, quaterChunkSize)
	require.NoError(t, err)
	assert.Equal(t, len(target), n)
	assert.Equal(t, target, source[quaterChunkSize:len(source)-quaterChunkSize])

	// Let's read some data from the middle of the stream < chunk size
	target = make([]byte, ChunkSize/4)
	n, err = asyncReader.readAt(target, ChunkSize+ChunkSize/4)
	require.NoError(t, err)
	assert.Equal(t, quaterChunkSize, n)
	assert.Equal(t, target, source[ChunkSize+quaterChunkSize:ChunkSize+quaterChunkSize*2])

	// Let's read all the data
	target = make([]byte, len(source))
	n, err = asyncReader.readAt(target, 0)
	require.NoError(t, err)
	assert.Equal(t, len(source), n)
	assert.Equal(t, target, source)

	// Let's try to read more data then available in the stream
	target = make([]byte, ChunkSize*2)
	n, err = asyncReader.readAt(target, ChunkSize*4)
	require.Error(t, err)
	assert.Equal(t, err, io.ErrUnexpectedEOF)
	assert.Equal(t, ChunkSize, n)
	assert.Equal(t, target[:ChunkSize/2], source[ChunkSize*4:]) // We read only last half chunk

	// Let's try to read data beyond the end of the stream
	target = make([]byte, ChunkSize*2)
	n, err = asyncReader.readAt(target, ChunkSize*5)
	require.Error(t, err)
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, 0, n)
}
