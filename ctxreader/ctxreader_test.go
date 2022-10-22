package ctxreader

import (
	"context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type testReader struct {
	closed bool
}

func (r *testReader) Read(p []byte) (int, error) {
	return rand.Reader.Read(p)
}

func (r *testReader) Close() error {
	r.closed = true
	return nil
}

type CtxReaderTestSuite struct {
	suite.Suite
}

func (s *CtxReaderTestSuite) TestReadUntilCanceled() {
	ctx, cancel := context.WithCancel(context.Background())

	r := New(ctx, &testReader{}, false)
	p := make([]byte, 1024)

	_, err := r.Read(p)
	require.Nil(s.T(), err)

	cancel()
	time.Sleep(time.Second)

	_, err = r.Read(p)
	require.Equal(s.T(), err, context.Canceled)
}

func (s *CtxReaderTestSuite) TestReturnOriginalOnBackgroundContext() {
	rr := &testReader{}
	r := New(context.Background(), rr, false)

	require.Equal(s.T(), rr, r)
}

func (s *CtxReaderTestSuite) TestClose() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rr := &testReader{}
	New(ctx, rr, true).Close()

	require.True(s.T(), rr.closed)
}

func (s *CtxReaderTestSuite) TestCloseOnCancel() {
	ctx, cancel := context.WithCancel(context.Background())

	rr := &testReader{}
	New(ctx, rr, true)

	cancel()
	time.Sleep(time.Second)

	require.True(s.T(), rr.closed)
}

func (s *CtxReaderTestSuite) TestDontCloseOnCancel() {
	ctx, cancel := context.WithCancel(context.Background())

	rr := &testReader{}
	New(ctx, rr, false)

	cancel()
	time.Sleep(time.Second)

	require.False(s.T(), rr.closed)
}

func TestCtxReader(t *testing.T) {
	suite.Run(t, new(CtxReaderTestSuite))
}

func BenchmarkRawReader(b *testing.B) {
	r := testReader{}

	b.ResetTimer()

	p := make([]byte, 1024)
	for i := 0; i < b.N; i++ {
		r.Read(p)
	}
}

func BenchmarkCtxReader(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	r := New(ctx, &testReader{}, true)

	b.ResetTimer()

	p := make([]byte, 1024)
	for i := 0; i < b.N; i++ {
		r.Read(p)
	}
}
