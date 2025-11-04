package abs

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/httprange"
)

// TestServer is a mock Azure Blob Storage server for testing
// It's a very light version of azurito
type TestServer struct {
	server         *httptest.Server
	stagedBlocks   map[string][]string          // container/blob -> blockIDs
	blockData      map[string]map[string][]byte // container/blob -> blockID -> data
	committedBlobs map[string][]byte            // container/blob -> committed data
	headers        map[string]http.Header       // container/blob -> HTTP headers (ETag, Last-Modified, Content-Type, etc.)
	mu             sync.Mutex
}

// NewAbsServer creates and starts a new mock Azure Blob Storage server
func NewAbsServer() (*TestServer, error) {
	abs := &TestServer{
		stagedBlocks:   make(map[string][]string),
		blockData:      make(map[string]map[string][]byte),
		committedBlobs: make(map[string][]byte),
		headers:        make(map[string]http.Header),
	}

	abs.server = httptest.NewTLSServer(http.HandlerFunc(abs.handler))

	return abs, nil
}

// handler handles Azure Blob Storage API requests
func (s *TestServer) handler(rw http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Parse path: /{container}/{blob}
	parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/"), "/", 2)
	if len(parts) < 2 {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	container := parts[0]
	blobName := parts[1]
	key := fmt.Sprintf("%s/%s", container, blobName)

	// Handle different Azure Blob Storage operations
	comp := r.URL.Query().Get("comp")
	blockID := r.URL.Query().Get("blockid")

	switch {
	case r.Method == http.MethodPut && comp == "block" && blockID != "":
		// StageBlock operation
		data, err := io.ReadAll(r.Body)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Initialize block data map if needed
		if s.blockData[key] == nil {
			s.blockData[key] = make(map[string][]byte)
		}

		// Store the block data
		s.blockData[key][blockID] = data

		// Track block ID in order
		if s.stagedBlocks[key] == nil {
			s.stagedBlocks[key] = []string{}
		}

		// Only add if not already present
		if !slices.Contains(s.stagedBlocks[key], blockID) {
			s.stagedBlocks[key] = append(s.stagedBlocks[key], blockID)
		}

		rw.WriteHeader(http.StatusCreated)

	case r.Method == http.MethodPut && comp == "blocklist":
		// CommitBlockList operation
		body, _ := io.ReadAll(r.Body)

		// Parse block IDs from XML (simplified - just extract blockid values)
		blockIDs := []string{}
		for _, id := range s.stagedBlocks[key] {
			if strings.Contains(string(body), id) {
				blockIDs = append(blockIDs, id)
			}
		}

		// Commit the blocks
		var result []byte
		for _, blockID := range blockIDs {
			if data, ok := s.blockData[key][blockID]; ok {
				result = append(result, data...)
			}
		}

		s.committedBlobs[key] = result

		// Store headers
		lastMod := time.Now().UTC()

		headers := make(http.Header)
		headers.Set(httpheaders.ContentType, r.Header.Get(httpheaders.ContentType))
		headers.Set(httpheaders.Etag, fmt.Sprintf(`"%x"`, md5.Sum(result)))
		headers.Set(httpheaders.LastModified, lastMod.Format(http.TimeFormat))
		headers.Set(httpheaders.ContentLength, fmt.Sprintf("%d", len(result)))

		s.headers[key] = headers

		// Clean up staged blocks
		delete(s.stagedBlocks, key)
		delete(s.blockData, key)

		rw.WriteHeader(http.StatusCreated)

	case r.Method == http.MethodPut && comp == "" && blockID == "":
		// Normal (non-partial) blob upload - PUT without block operations
		data, err := io.ReadAll(r.Body)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Store the blob data directly
		s.committedBlobs[key] = data

		// Store headers
		lastMod := time.Now().UTC()

		headers := make(http.Header)
		etag := fmt.Sprintf(`"%x"`, md5.Sum(data))
		headers.Set(httpheaders.Etag, etag)
		headers.Set(httpheaders.LastModified, lastMod.Format(http.TimeFormat))
		headers.Set(httpheaders.ContentType, r.Header.Get(httpheaders.ContentType))
		headers.Set(httpheaders.ContentLength, fmt.Sprintf("%d", len(data)))
		s.headers[key] = headers

		// Set response headers
		rw.Header().Set(httpheaders.Etag, etag)
		rw.Header().Set(httpheaders.LastModified, lastMod.Format(http.TimeFormat))

		rw.WriteHeader(http.StatusCreated)

	case r.Method == http.MethodDelete:
		// Delete blob operation
		delete(s.committedBlobs, key)
		delete(s.stagedBlocks, key)
		delete(s.blockData, key)
		delete(s.headers, key)
		rw.WriteHeader(http.StatusAccepted)

	case r.Method == http.MethodGet:
		// Get blob operation
		data, ok := s.committedBlobs[key]
		if !ok {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<Error>
  <Code>BlobNotFound</Code>
  <Message>The specified blob does not exist.</Message>
</Error>`))
			return
		}

		// Get stored headers
		headers := s.headers[key].Clone()

		// Handle range requests - Azure uses x-ms-range header
		rangeHeader := r.Header.Get("x-ms-range")
		if rangeHeader != "" {
			headers.Del(httpheaders.ContentLength)
		}
		httpheaders.CopyAll(headers, rw.Header(), true)

		rw.Header().Set(httpheaders.AcceptRanges, "bytes")

		if rangeHeader == "" {
			// Full content
			rw.Header().Set(httpheaders.ContentLength, fmt.Sprintf("%d", len(data)))
			rw.WriteHeader(http.StatusOK)
			rw.Write(data)
			return
		}

		// Parse range header
		start, end, err := httprange.Parse(rangeHeader)
		if err != nil {
			rw.Header().Set(httpheaders.ContentRange, fmt.Sprintf("bytes */%d", len(data)))
			rw.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
			return
		}

		// Handle open-ended range (e.g., "bytes=0-")
		if end == -1 {
			end = int64(len(data)) - 1
		}

		// Validate range
		if start < 0 || start >= int64(len(data)) || end >= int64(len(data)) || start > end {
			rw.Header().Set(httpheaders.ContentRange, fmt.Sprintf("bytes */%d", len(data)))
			rw.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
			return
		}

		// Serve partial content
		rangeData := data[start : end+1]
		rw.Header().Set(httpheaders.ContentLength, fmt.Sprintf("%d", len(rangeData)))
		rw.Header().Set(httpheaders.ContentRange, fmt.Sprintf("bytes %d-%d/%d", start, end, len(data)))
		rw.WriteHeader(http.StatusPartialContent)
		rw.Write(rangeData)

	default:
		rw.WriteHeader(http.StatusNotImplemented)
	}
}

// Close stops the server
func (s *TestServer) Close() {
	s.server.Close()
}

// URL returns the server URL
func (s *TestServer) URL() string {
	return s.server.URL
}
