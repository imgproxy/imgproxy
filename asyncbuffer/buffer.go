// Package asyncbuffer provides an asynchronous buffer that reads data from an
// io.Reader in the background.
//
// When created, AsyncBuffer starts reading from the upstream reader in the
// background. If a read error occurs, it is stored and can be checked with
// AsyncBuffer.Error().
//
// When reading through AsyncBuffer.Reader().Read(), the error is only returned
// once the reader reaches the point where the error occurred. In other words,
// errors are delayed until encountered by the reader.
//
// However, AsyncBuffer.Close() and AsyncBuffer.Error() will immediately return
// any stored error, even if the reader has not yet reached the error point.
package asyncbuffer

import (
	"errors"
	"io"
	"sync"
	"sync/atomic"
)

// ChunkSize is the size of each chunk in bytes
const ChunkSize = 4096

// byteChunk is a struct that holds a buffer and the data read from the upstream reader
// data slice is required since the chunk read may be smaller than ChunkSize
type byteChunk struct {
	buf  []byte
	data []byte
}

// chunkPool is a global sync.Pool that holds byteChunk objects for
// all readers
var chunkPool = sync.Pool{
	New: func() any {
		buf := make([]byte, ChunkSize)

		return &byteChunk{
			buf:  buf,
			data: buf[:0],
		}
	},
}

// AsyncBuffer is a wrapper around io.Reader that reads data in chunks
// in background and allows reading from synchronously.
type AsyncBuffer struct {
	r io.Reader // Upstream reader

	chunks []*byteChunk // References to the chunks read from the upstream reader

	err      atomic.Value // Error that occurred during reading
	finished atomic.Bool  // Indicates that the buffer has finished reading
	len      atomic.Int64 // Total length of the data read
	closed   atomic.Bool  // Indicates that the buffer was closed

	mu             sync.RWMutex  // Mutex on chunks slice
	newChunkSignal chan struct{} // Tick-tock channel that indicates that a new chunk is ready
}

// Underlying Reader that provides io.ReadSeeker interface for the actual data reading
// What is the purpose of this Reader?
type Reader struct {
	ab  *AsyncBuffer
	pos int64
}

// FromReadCloser creates a new AsyncBuffer that reads from the given io.Reader in background
func FromReader(r io.Reader) *AsyncBuffer {
	ab := &AsyncBuffer{
		r:              r,
		newChunkSignal: make(chan struct{}),
	}

	go ab.readChunks()

	return ab
}

// getNewChunkSignal returns the channel that signals when a new chunk is ready
// Lock is required to read the channel, so it is not closed while reading
func (ab *AsyncBuffer) getNewChunkSignal() chan struct{} {
	ab.mu.RLock()
	defer ab.mu.RUnlock()

	return ab.newChunkSignal
}

// addChunk adds a new chunk to the AsyncBuffer, increments len and signals that a chunk is ready
func (ab *AsyncBuffer) addChunk(chunk *byteChunk) {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	if ab.closed.Load() {
		// If the buffer is closed, then the chunk is not needed anymore,
		// so we return it to the pool
		chunkPool.Put(chunk)
		return
	}

	// Store the chunk, increase chunk size, increase length of the data read
	ab.chunks = append(ab.chunks, chunk)
	ab.len.Add(int64(len(chunk.data)))

	// Signal that a chunk is ready
	currSignal := ab.newChunkSignal
	ab.newChunkSignal = make(chan struct{})
	close(currSignal)
}

// readChunks reads data from the upstream reader in background and stores them in the pool
func (ab *AsyncBuffer) readChunks() {
	defer func() {
		// Indicate that the buffer has finished reading
		ab.finished.Store(true)

		// Since the buffer is finished reading, we should close the channel to unlock any waiting readers
		close(ab.newChunkSignal)
	}()

	// Stop reading if the buffer is closed
	for !ab.closed.Load() {
		// Get a chunk from the pool
		// If the pool is empty, it will create a new byteChunk with ChunkSize
		chunk, ok := chunkPool.Get().(*byteChunk)
		if !ok {
			ab.err.Store(errors.New("asyncbuffer.AsyncBuffer.readChunks: failed to get chunk from pool"))
			return
		}

		// Read data into the chunk's buffer
		n, err := io.ReadFull(ab.r, chunk.buf)

		// If it's not the EOF, we need to store the error
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			ab.err.Store(err)
			return
		}

		// No bytes were read (n == 0), we can return the chunk to the pool
		if err == io.EOF || n == 0 {
			chunkPool.Put(chunk)
			return
		}

		// Resize the chunk's data slice to the number of bytes read
		chunk.data = chunk.buf[:n]

		// Store the reference to the chunk in the AsyncBuffer
		ab.addChunk(chunk)

		// We got ErrUnexpectedEOF meaning that some bytes were read, but this is the
		// end of the stream, so we can stop reading
		if err == io.ErrUnexpectedEOF {
			return
		}
	}
}

// closedError returns an error if the attempt to read on a closed reader was made.
// If the reader had an error, it returns that error instead.
func (ab *AsyncBuffer) closedError() error {
	// If the reader is closed, we return the error or nil
	if ab.closed.Load() {
		err := ab.Error()
		if err == nil {
			err = errors.New("asyncbuffer.AsyncBuffer.ReadAt: attempt to read on closed reader")
		}

		return err
	}

	return nil
}

// offsetAvailable checks if the data at the given offset is available for reading.
// It may return io.EOF if the reader is finished reading and the offset is beyond the end of the stream.
func (ab *AsyncBuffer) offsetAvailable(off int64) (bool, error) {
	// We can not read data from the closed reader, none
	if err := ab.closedError(); err != nil {
		return false, err
	}

	// In case the offset falls within the already read chunks, we can return immediately,
	// even if error has occurred in the future
	if off < ab.len.Load() {
		return true, nil
	}

	// In case the reader is finished reading, and we have not read enough
	// data yet, return either error or EOF
	if ab.finished.Load() {
		// In case, error has occurred, we need to return it
		err := ab.Error()
		if err != nil {
			return false, err
		}

		// Otherwise, it's EOF if the offset is beyond the end of the stream
		return false, io.EOF
	}

	// No available data
	return false, nil
}

// WaitFor waits for the data to be ready at the given offset. nil means ok.
// It guarantees that the chunk at the given offset is ready to be read.
func (ab *AsyncBuffer) WaitFor(off int64) error {
	for {
		ok, err := ab.offsetAvailable(off)
		if ok || err != nil {
			return err
		}

		<-ab.getNewChunkSignal()
	}
}

// Wait waits for the reader to finish reading all data and returns
// the total length of the data read.
func (ab *AsyncBuffer) Wait() (int64, error) {
	for {
		// We can not read data from the closed reader even if there were no errors
		if err := ab.closedError(); err != nil {
			return 0, err
		}

		// In case the reader is finished reading, we can return immediately
		if ab.finished.Load() {
			// If there was an error during reading, we need to return it no matter what position
			// had the error happened
			return ab.len.Load(), ab.Error()
		}

		// Lock until the next chunk is ready
		<-ab.getNewChunkSignal()
	}
}

// Error returns the error that occurred during reading data in background.
func (ab *AsyncBuffer) Error() error {
	err := ab.err.Load()
	if err == nil {
		return nil
	}

	errCast, ok := err.(error)
	if !ok {
		return errors.New("asyncbuffer.AsyncBuffer.Error: failed to get error")
	}

	return errCast
}

// readChunkAt copies data from the chunk at the given absolute offset to the provided slice.
// Chunk must be available when this method is called.
// Returns the number of bytes copied to the slice or 0 if chunk has no data
// (eg. offset is beyond the end of the stream).
func (ab *AsyncBuffer) readChunkAt(p []byte, off int64) int {
	// If the chunk is not available, we return 0
	if off >= ab.len.Load() {
		return 0
	}

	ind := off / ChunkSize // chunk index
	chunk := ab.chunks[ind]

	startOffset := off % ChunkSize // starting offset in the chunk

	// If the offset in current chunk is greater than the data
	// it has, we return 0
	if startOffset >= int64(len(chunk.data)) {
		return 0
	}

	// Copy data to the target slice. The number of bytes to copy is limited by the
	// size of the target slice and the size of the data in the chunk.
	return copy(p, chunk.data[startOffset:])
}

// readAt reads data from the AsyncBuffer at the given offset.
//
// If full is true:
//
// The behaviour is similar to io.ReaderAt.ReadAt. It blocks until the maxumum amount of data possible
// is read from the buffer. It may return io.UnexpectedEOF in case we tried to read more data than was
// available in the buffer.
//
// If full is false:
//
// It behaves like a regular non-blocking Read.
func (ab *AsyncBuffer) readAt(p []byte, off int64) (int, error) {
	size := int64(len(p)) // total size of the data to read

	if off < 0 {
		return 0, errors.New("asyncbuffer.AsyncBuffer.readAt: negative offset")
	}

	// Wait for the offset to be available.
	// It may return io.EOF if the offset is beyond the end of the stream.
	err := ab.WaitFor(off)
	if err != nil {
		return 0, err
	}

	ab.mu.RLock()
	defer ab.mu.RUnlock()

	// If the reader is closed, we return an error
	if err := ab.closedError(); err != nil {
		return 0, err
	}

	// Read data from the first chunk
	n := ab.readChunkAt(p, off)
	if n == 0 {
		return 0, io.EOF // Failed to read any data: means we tried to read beyond the end of the stream
	}

	size -= int64(n)
	off += int64(n) // Here and beyond off always points to the last read byte + 1

	// Now, let's try to read the rest of the data from next chunks while they are available
	for size > 0 {
		// If data is not available at the given offset, we can return data read so far.
		ok, err := ab.offsetAvailable(off)
		if !ok || err != nil {
			return n, err
		}

		// Read data from the next chunk
		nX := ab.readChunkAt(p[n:], off)
		n += nX
		size -= int64(nX)
		off += int64(nX)

		// If we read data shorter than ChunkSize or, in case that was the last chunk, less than
		// the size of the tail, return kind of EOF
		if int64(nX) < min(size, int64(ChunkSize)) {
			return n, io.EOF
		}
	}

	return n, nil
}

// Close closes the AsyncBuffer and releases all resources.
// It returns an error if the reader was already closed or if there was
// an error during reading data in background even if none of the subsequent
// readers have reached the position where the error occurred.
func (ab *AsyncBuffer) Close() error {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	// If the reader is already closed, we return immediately error or nil
	if ab.closed.Load() {
		return ab.Error()
	}

	ab.closed.Store(true)

	// Return all chunks to the pool
	for _, chunk := range ab.chunks {
		chunkPool.Put(chunk)
	}

	return nil
}

// Reader returns an io.ReadSeeker+io.ReaderAt that can be used to read actual data from the AsyncBuffer
func (ab *AsyncBuffer) Reader() *Reader {
	return &Reader{ab: ab, pos: 0}
}

// Read reads data from the AsyncBuffer.
func (r *Reader) Read(p []byte) (int, error) {
	n, err := r.ab.readAt(p, r.pos)
	if err == nil {
		r.pos += int64(n)
	}

	return n, err
}

// Seek sets the position of the reader to the given offset and returns the new position
func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		r.pos = offset

	case io.SeekCurrent:
		r.pos += offset

	case io.SeekEnd:
		size, err := r.ab.Wait()
		if err != nil {
			return 0, err
		}

		r.pos = size + offset

	default:
		return 0, errors.New("asyncbuffer.AsyncBuffer.ReadAt: invalid whence")
	}

	if r.pos < 0 {
		return 0, errors.New("asyncbuffer.AsyncBuffer.ReadAt: negative position")
	}

	return r.pos, nil
}
