package bufreader

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

// BufferedReaderTestSuite defines the test suite for the buffered reader
type BufferedReaderTestSuite struct {
	suite.Suite
}

func (s *BufferedReaderTestSuite) TestRead() {
	data := "hello world"
	br := New(strings.NewReader(data))

	// First read
	p1 := make([]byte, 5)
	n1, err1 := br.Read(p1)
	s.Require().NoError(err1)
	s.Equal(5, n1)
	s.Equal("hello", string(p1))

	// Second read
	p2 := make([]byte, 6)
	n2, err2 := br.Read(p2)
	s.Require().NoError(err2)
	s.Equal(6, n2)
	s.Equal(" world", string(p2))

	// Verify position
	s.Equal(11, br.pos)
}

func (s *BufferedReaderTestSuite) TestEOF() {
	data := "hello"
	br := New(strings.NewReader(data))

	// Read all data
	p1 := make([]byte, 5)
	n1, err1 := br.Read(p1)
	s.Require().NoError(err1)
	s.Equal(5, n1)
	s.Equal("hello", string(p1))

	// Try to read more - should get EOF
	p2 := make([]byte, 5)
	n2, err2 := br.Read(p2)
	s.Equal(io.EOF, err2)
	s.Equal(0, n2)
}

func (s *BufferedReaderTestSuite) TestEOF_WhenDataExhausted() {
	data := "hello"
	br := New(strings.NewReader(data))

	// Try to read more than available
	p := make([]byte, 10)
	n, err := br.Read(p)

	s.Require().NoError(err)
	s.Equal(5, n)
	s.Equal("hello", string(p[:n]))
}

func (s *BufferedReaderTestSuite) TestPeek() {
	data := "hello world"
	br := New(strings.NewReader(data))

	// Peek at first 5 bytes
	peeked, err := br.Peek(5)
	s.Require().NoError(err)
	s.Equal("hello", string(peeked))
	s.Equal(0, br.pos) // Position should not change

	// Read the same data to verify peek didn't consume it
	p := make([]byte, 5)
	n, err := br.Read(p)
	s.Require().NoError(err)
	s.Equal(5, n)
	s.Equal("hello", string(p))
	s.Equal(5, br.pos) // Position should now be updated

	// Peek at the next 7 bytes (which are beyond the EOF)
	peeked2, err := br.Peek(7)
	s.Require().NoError(err)
	s.Equal(" world", string(peeked2))
	s.Equal(5, br.pos) // Position should still be 5
}

// TestBufferedReaderSuite runs the test suite
func TestBufferedReader(t *testing.T) {
	suite.Run(t, new(BufferedReaderTestSuite))
}
