package ctxreader

import (
	"context"
	"io"
	"sync"
	"sync/atomic"
)

type ctxReader struct {
	r         io.ReadCloser
	err       atomic.Value
	closeOnce sync.Once
}

func (r *ctxReader) Read(p []byte) (int, error) {
	if err := r.err.Load(); err != nil {
		return 0, err.(error)
	}
	return r.r.Read(p)
}

func (r *ctxReader) Close() (err error) {
	r.closeOnce.Do(func() { err = r.r.Close() })
	return
}

func New(ctx context.Context, r io.ReadCloser, closeOnDone bool) io.ReadCloser {
	if ctx.Done() == nil {
		return r
	}

	ctxr := ctxReader{r: r}

	go func(ctx context.Context) {
		<-ctx.Done()
		ctxr.err.Store(ctx.Err())
		if closeOnDone {
			ctxr.closeOnce.Do(func() { ctxr.r.Close() })
		}
	}(ctx)

	return &ctxr
}
