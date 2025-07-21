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

	chunks      []*byteChunk // References to the chunks read from the upstream reader
	chunksReady atomic.Int64 // Number of chunks that were read

	err      atomic.Value // Error that occurred during reading
	finished atomic.Bool  // Indicates that the reader has finished reading
	len      atomic.Int64 // Total length of the data read
	closed   atomic.Bool  // Indicates that the reader was closed

	mu             sync.RWMutex  // Mutex on chunks slice
	newChunkSignal chan struct{} // Tick-tock channel that indicates that a new chunk is ready
}

// Underlying Reader that provides io.ReadSeeker interface for the actual data reading
// What is the purpose of this Reader?
type Reader struct {
	io.ReadSeeker
	io.ReaderAt
	ar  *AsyncBuffer
	pos int64
}

// NewAsyncBuffer creates a new AsyncBuffer that reads from the given io.Reader in background
func NewAsyncBuffer(r io.Reader) *AsyncBuffer {
	ar := &AsyncBuffer{
		r:              r,
		newChunkSignal: make(chan struct{}),
	}

	go ar.readChunks()

	return ar
}

// getNewChunkSignal returns the channel that signals when a new chunk is ready
// Lock is required to read the channel, so it is not closed while reading
func (ar *AsyncBuffer) getNewChunkSignal() chan struct{} {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	return ar.newChunkSignal
}

// addChunk adds a new chunk to the AsyncBuffer, increments len and signals that a chunk is ready
func (ar *AsyncBuffer) addChunk(chunk *byteChunk) {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	// Store the chunk, increase chunk size, increase length of the data read
	ar.chunks = append(ar.chunks, chunk)
	ar.chunksReady.Add(1)
	ar.len.Add(int64(len(chunk.data)))

	// Signal that a chunk is ready
	currSignal := ar.newChunkSignal
	ar.newChunkSignal = make(chan struct{})
	close(currSignal)
}

// finish marks the reader as finished
func (ar *AsyncBuffer) finish() {
	// Indicate that the reader has finished reading
	ar.finished.Store(true)

	// This indicates that Close() was called before all the chunks were read, we do not need to close the channel
	// since it was closed already.
	if !ar.closed.Load() {
		close(ar.newChunkSignal)
	}
}

// readChunks reads data from the upstream reader in background and stores them in the pool
func (ar *AsyncBuffer) readChunks() {
	defer ar.finish()

	// Stop reading if the reader is finished
	for !ar.finished.Load() {
		// Get a chunk from the pool
		// If the pool is empty, it will create a new byteChunk with ChunkSize
		chunk, ok := chunkPool.Get().(*byteChunk)
		if !ok {
			ar.err.Store(errors.New("asyncbuffer.AsyncBuffer.readChunks: failed to get chunk from pool"))
			return
		}

		// Read data into the chunk's buffer
		n, err := io.ReadFull(ar.r, chunk.buf)

		// If it's not the EOF, we need to store the error
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			ar.err.Store(err)
			return
		}

		// No bytes were read (n == 0), we can return the chunk to the pool
		if err == io.EOF {
			chunkPool.Put(chunk)
			return
		}

		// Resize the chunk's data slice to the number of bytes read
		chunk.data = chunk.buf[:n]

		// Store the reference to the chunk in the AsyncBuffer
		ar.addChunk(chunk)

		// We got ErrUnexpectedEOF meaning that some bytes were read, but this is the
		// end of the stream, so we can stop reading
		if err == io.ErrUnexpectedEOF {
			return
		}
	}
}

// offsetAvailable checks if the data at the given offset is available for reading.
// It may return io.EOF if the reader is finished reading and the offset is beyond the end of the stream.
func (ar *AsyncBuffer) offsetAvailable(off int64) (bool, error) {
	// We can not read data from the closed reader, none
	if ar.closed.Load() {
		return false, ar.error()
	}

	// In case the offset falls within the already read chunks, we can return immediately,
	// even if error has occurred in the future
	if off < ar.chunksReady.Load()*ChunkSize {
		return true, nil
	}

	// In case the reader is finished reading, and we have not read enough
	// data yet, return either error or EOF
	if ar.finished.Load() {
		// In case, error has occurred, we need to return it
		err := ar.error()
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
func (ar *AsyncBuffer) WaitFor(off int64) error {
	for {
		ok, err := ar.offsetAvailable(off)
		if ok || err != nil {
			return err
		}

		<-ar.getNewChunkSignal()
	}
}

// Wait waits for the reader to finish reading all data
func (ar *AsyncBuffer) Wait() error {
	for {
		// We can not read data from the closed reader even if there were no errors
		if ar.closed.Load() {
			return ar.error()
		}

		// In case the reader is finished reading, we can return immediately
		if ar.finished.Load() {
			// If there was an error during reading, we need to return it no matter what position
			// had the error happened
			err := ar.err.Load()
			if err != nil {
				err, ok := err.(error)
				if !ok {
					return errors.New("asyncbuffer.AsyncBuffer.Wait: failed to get error")
				}

				return err
			}

			return nil
		}

		// Lock until the next chunk is ready
		<-ar.getNewChunkSignal()
	}
}

// size returns the total size of the data read. In case of an error happened during reading or
// the reader has not finished yet, it returns -1
func (ar *AsyncBuffer) size() (int64, error) {
	err := ar.error()
	if err != nil {
		return -1, err
	}

	if ar.finished.Load() {
		return ar.len.Load(), nil
	}

	return -1, nil
}

// error returns the error that occurred during reading
func (ar *AsyncBuffer) error() error {
	err := ar.err.Load()
	if err == nil {
		return nil
	}

	errCast, ok := err.(error)
	if !ok {
		return errors.New("asyncbuffer.AsyncBuffer.Error: failed to get error")
	}

	return errCast
}

// readAt reads data from the AsyncBuffer at the given offset.
//
// If full is true:
//
// The behaviour is similar to io.ReaderAt.ReadAt. It blocks until the maxumum amount of data possible
// is read from the buffer. It may return io.UnexpectedEOF if the requested amount of
// data is not available in the buffer.
//
// If full is false:
//
// It behaves like a regular Read.
func (ar *AsyncBuffer) readAt(p []byte, off int64, full bool) (int, error) {
	if off < 0 {
		return 0, errors.New("asyncbuffer.AsyncBuffer.readAt: negative offset")
	}

	// Check if the offset is available/wait for it to be available.
	// It may return io.EOF if the offset is beyond the end of the stream.
	if full {
		err := ar.WaitFor(off)
		if err != nil {
			return 0, err
		}
	} else {
		ok, err := ar.offsetAvailable(off)
		if !ok || err != nil {
			return 0, err
		}
	}

	size := int64(len(p))       // total size of the data to read
	chunkInd := off / ChunkSize // number of full chunks of starting offset
	chunkOff := off % ChunkSize // starting offset in the first chunk

	// If the chunk index is out of bounds, we return EOF
	if chunkInd >= ar.chunksReady.Load() {
		return 0, io.EOF
	}

	// We read the actual chunk data, so we need to prevent ar.chunks from being modified
	ar.mu.RLock()
	chunk := ar.chunks[chunkInd]
	ar.mu.RUnlock()

	// If the offset in current chunk is greater that the data
	// it has (that is the last chunk), we return EOF
	if chunkOff >= int64(len(chunk.data)) {
		return 0, io.EOF
	}

	// how many bytes we need to read from the first chunk (can be less than ChunkSize)
	shouldReadLen := ChunkSize - chunkOff
	shouldReadLen = min(shouldReadLen, size)

	// Copy the data from the chunk to the slice
	n := copy(p, chunk.data[chunkOff:shouldReadLen+chunkOff])
	if n == 0 {
		return 0, io.EOF
	}

	chunkInd += 1
	size -= shouldReadLen

	// Now, let's try to read the rest of the data from next chunks until they are available
	for size > 0 {
		// Let's wait for the next chunk to be ready or check if it is available
		// This method may return io.EOF if the reader is finished reading and the offset is beyond the end of the stream.
		if full {
			err := ar.WaitFor(chunkInd * ChunkSize)

			if err != nil {
				if err == io.EOF {
					return n, io.ErrUnexpectedEOF // Means that we expected more data, but it is not available
				} else {
					return n, err
				}
			}
		} else {
			// For blocking version, we need to return this EOF and the data read so far
			ok, err := ar.offsetAvailable(chunkInd * ChunkSize)
			if !ok || err != nil {
				return n, err
			}
		}

		// Read the actual chunk data
		ar.mu.RLock()
		chunk := ar.chunks[chunkInd]
		ar.mu.RUnlock()

		shouldReadLen = min(ChunkSize, size)
		shouldReadLen = min(shouldReadLen, int64(len(chunk.data)))

		// Append next chunk data to the slice
		nAvailable := copy(p[n:], chunk.data[0:shouldReadLen])

		// In case we failed to read enough data from the next chunk, we return EOF
		if nAvailable < int(shouldReadLen) {
			if full {
				return n, io.ErrUnexpectedEOF // It is unexpected EOF, we expected more data, but it is not available
			} else {
				return n, io.EOF // We reached the end of the stream, but we read all the data that was available at this moment
			}
		}

		n += nAvailable
		size -= int64(shouldReadLen)
		chunkInd += 1
	}

	return n, nil
}

// Close closes the AsyncBuffer and releases all resources
func (ar *AsyncBuffer) Close() {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	// If the reader is already closed, we return immediately
	if ar.closed.Load() {
		return
	}

	// All the methods that can be called on the AsyncBuffer should return an error
	ar.err.Store(errors.New("asyncbuffer.AsyncBuffer.ReadAt: attempt to read on closed reader"))
	ar.closed.Store(true)

	// If the reader is still running, we need to signal that it should stop and close the channel
	if !ar.finished.Load() {
		ar.finished.Store(true)
		close(ar.newChunkSignal)
	}

	// Return all chunks to the pool
	for _, chunk := range ar.chunks {
		chunkPool.Put(chunk)
	}
}

// Reader returns an io.ReadSeeker+io.ReaderAt that can be used to read actual data from the AsyncBuffer
func (ar *AsyncBuffer) Reader() *Reader {
	return &Reader{ar: ar, pos: 0}
}

// ReadAt reads data from the AsyncBuffer at the given offset.
// The method behaves exactly like io.ReaderAt.ReadAt.
// It blocks until all the data is ready.
func (r Reader) ReadAt(p []byte, off int64) (int, error) {
	return r.ar.readAt(p, off, true)
}

// Size returns the total size of the data read by the AsyncBuffer or error,
// -1 if the reader has not finished yet
func (r Reader) Size() (int64, error) {
	return r.ar.size()
}

// ReadFull reads data from the AsyncBuffer in sync mode.
func (r *Reader) ReadFull(p []byte) (int, error) {
	n, err := r.ar.readAt(p, r.pos, true)
	if err != nil {
		return n, err
	}

	r.pos += int64(n)

	return n, nil
}

// ReadFull reads data from the AsyncBuffer in async mode.
func (r *Reader) Read(p []byte) (int, error) {
	n, err := r.ar.readAt(p, r.pos, false)
	if err != nil {
		return n, err
	}

	r.pos += int64(n)

	return n, nil
}

// Seek sets the position of the reader to the given offset and returns the new position
func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		if err := r.ar.WaitFor(offset); err != nil {
			return 0, err
		}
		r.pos = offset

	case io.SeekCurrent:
		if err := r.ar.WaitFor(r.pos + offset); err != nil {
			return 0, err
		}
		r.pos += offset

	case io.SeekEnd:
		err := r.ar.Wait()
		if err != nil {
			return 0, err
		}

		size, err := r.ar.size()
		if err != nil {
			return -1, err
		}

		r.pos = size + offset

	default:
		return 0, errors.New("asyncbuffer.AsyncBuffer.ReadAt: invalid whence")
	}

	if r.pos < 0 {
		return 0, errors.New("asyncbuffer.AsyncBuffer.ReadAt: negative position")
	}

	size, err := r.ar.size()
	if err != nil {
		return -1, err
	}

	if r.pos > size {
		return -1, io.EOF
	}

	return r.pos, nil
}
