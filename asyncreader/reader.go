package asyncreader

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

// AsyncReader is a wrapper around io.Reader that reads data in chunks
// in background and allows reading from synchronously.
type AsyncReader struct {
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
	ar  *AsyncReader
	pos int64
}

// NewAsyncReader creates a new AsyncReader that reads from the given io.Reader in background
func NewAsyncReader(r io.Reader) *AsyncReader {
	ar := &AsyncReader{
		r:              r,
		newChunkSignal: make(chan struct{}),
	}

	go ar.readChunks()

	return ar
}

// getNewChunkSignal returns the channel that signals when a new chunk is ready
// Lock is required to read the channel, so it is not closed while reading
func (ar *AsyncReader) getNewChunkSignal() chan struct{} {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	return ar.newChunkSignal
}

// addChunk adds a new chunk to the AsyncReader, increments len and signals that a chunk is ready
func (ar *AsyncReader) addChunk(chunk *byteChunk) {
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
func (ar *AsyncReader) finish() {
	// Indicate that the reader has finished reading
	ar.finished.Store(true)

	// This indicates that Close() was called before all the chunks were read, we do not need to close the channel
	// since it was closed already.
	if !ar.closed.Load() {
		close(ar.newChunkSignal)
	}
}

// readChunks reads data from the upstream reader in background and stores them in the pool
func (ar *AsyncReader) readChunks() {
	defer ar.finish()

	// Stop reading if the reader is finished
	for !ar.finished.Load() {
		// Get a chunk from the pool
		// If the pool is empty, it will create a new byteChunk with ChunkSize
		chunk, ok := chunkPool.Get().(*byteChunk)
		if !ok {
			ar.err.Store(errors.New("asyncreader.AsyncReader.readChunks: failed to get chunk from pool"))
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

		// Store the reference to the chunk in the AsyncReader
		ar.addChunk(chunk)

		// We got ErrUnexpectedEOF meaning that some bytes were read, but this is the
		// end of the stream, so we can stop reading
		if err == io.ErrUnexpectedEOF {
			return
		}
	}
}

// WaitFor waits for the data to be ready at the given offset. nil means ok.
// It guarantees that the chunk at the given offset is ready to be read.
func (ar *AsyncReader) WaitFor(off int64) error {
	for {
		// In case the offset falls within the already read chunks, we can return immediately
		if off < ar.chunksReady.Load()*ChunkSize {
			return nil
		}

		// In case, error has occurred, we need to return it
		err := ar.error()
		if err != nil {
			return err
		}

		// In case the reader is finished reading, and we have not read enough
		// data yet, return EOF
		if ar.finished.Load() {
			return io.EOF
		}

		<-ar.getNewChunkSignal()
	}
}

// Wait waits for the reader to finish reading all data
func (ar *AsyncReader) Wait() error {
	for {
		if ar.finished.Load() {
			err := ar.err.Load()
			if err != nil {
				err, ok := err.(error)
				if !ok {
					return errors.New("asyncreader.AsyncReader.Wait: failed to get error")
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
func (ar *AsyncReader) size() (int64, error) {
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
func (ar *AsyncReader) error() error {
	err := ar.err.Load()
	if err == nil {
		return nil
	}

	errCast, ok := err.(error)
	if !ok {
		return errors.New("asyncreader.AsyncReader.Error: failed to get error")
	}

	return errCast
}

// readAt reads data from the AsyncReader at the given offset.
// Check io.ReaderAt interface.
//
// It has exactly the same behaviour as ReadFull:
// 1. It blocks if the data is not yet available at the given offset.
// 2. It returns io.UnexpectedEOF if n < len(p) and the end of the stream is reached.
func (ar *AsyncReader) readAt(p []byte, off int64) (int, error) {
	if off < 0 {
		return 0, errors.New("asyncreader.AsyncReader.ReadAt: negative offset")
	}

	// Wait for the chunk to be ready.
	// We should ensure that we have at least one byte available at the offset.
	err := ar.WaitFor(off + 1)
	if err != nil {
		return 0, err
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
	lenToRead := ChunkSize - chunkOff
	lenToRead = min(lenToRead, size)

	// Copy the data from the chunk to the slice
	n := copy(p, chunk.data[chunkOff:lenToRead+chunkOff])
	if n == 0 {
		return 0, io.EOF
	}

	chunkInd += 1
	size -= lenToRead

	for size > 0 {
		// Let's wait for the next chunk to be ready
		err := ar.WaitFor(chunkInd * ChunkSize)
		if err != nil && err != io.EOF {
			return n, err
		}

		// If the next chunk index is out of bounds, we return UnexpectedEOF
		if chunkInd >= int64(ar.chunksReady.Load()) {
			return n, io.ErrUnexpectedEOF
		}

		// Read the actual chunk data
		ar.mu.RLock()
		chunk := ar.chunks[chunkInd]
		ar.mu.RUnlock()

		lenToRead = min(ChunkSize, size)
		lenToRead = min(lenToRead, int64(len(chunk.data)))

		// Append next chunk data to the slice
		nextN := copy(p[n:], chunk.data[0:lenToRead])

		// In case we failed to read enough data from the next chunk, we return EOF
		if nextN < int(lenToRead) {
			return n, io.ErrUnexpectedEOF
		}

		n += nextN
		size -= int64(lenToRead)
		chunkInd += 1
	}

	return n, nil
}

// Close closes the AsyncReader and releases all resources
func (ar *AsyncReader) Close() {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	// If the reader is already closed, we return immediately
	if ar.closed.Load() {
		return
	}

	// All the methods that can be called on the AsyncReader should return an error
	ar.err.Store(errors.New("asyncreader.AsyncReader.ReadAt: attempt to read on closed reader"))
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

// Reader returns an io.ReadSeeker+io.ReaderAt that can be used to read actual data from the AsyncReader
func (ar *AsyncReader) Reader() *Reader {
	return &Reader{ar: ar, pos: 0}
}

// ReadAt reads data from the AsyncReader at the given offset
func (r Reader) ReadAt(p []byte, off int64) (int, error) {
	return r.ar.readAt(p, off)
}

// Size returns the total size of the data read by the AsyncReader or error,
// -1 if the reader has not finished yet
func (r Reader) Size() (int64, error) {
	return r.ar.size()
}

// Read reads data from the AsyncReader
func (r *Reader) Read(p []byte) (int, error) {
	if err := r.ar.WaitFor(r.pos); err != nil {
		return 0, err
	}

	n, err := r.ar.readAt(p, r.pos)
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
		return 0, errors.New("asyncreader.AsyncReader.ReadAt: invalid whence")
	}

	if r.pos < 0 {
		return 0, errors.New("asyncreader.AsyncReader.ReadAt: negative position")
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
