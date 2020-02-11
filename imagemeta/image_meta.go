package imagemeta

import (
	"bufio"
	"errors"
	"io"
	"sync"
	"sync/atomic"
)

type Meta interface {
	Format() string
	Width() int
	Height() int
}

type DecodeMetaFunc func(io.Reader) (Meta, error)

type meta struct {
	format        string
	width, height int
}

func (m *meta) Format() string {
	return m.format
}

func (m *meta) Width() int {
	return m.width
}

func (m *meta) Height() int {
	return m.height
}

type format struct {
	magic      string
	decodeMeta DecodeMetaFunc
}

type reader interface {
	io.Reader
	Peek(int) ([]byte, error)
}

var (
	formatsMu     sync.Mutex
	atomicFormats atomic.Value

	ErrFormat = errors.New("unknown image format")
)

func asReader(r io.Reader) reader {
	if rr, ok := r.(reader); ok {
		return rr
	}
	return bufio.NewReader(r)
}

func matchMagic(magic string, b []byte) bool {
	if len(magic) != len(b) {
		return false
	}
	for i, c := range b {
		if magic[i] != c && magic[i] != '?' {
			return false
		}
	}
	return true
}

func RegisterFormat(magic string, decodeMeta DecodeMetaFunc) {
	formatsMu.Lock()
	defer formatsMu.Unlock()

	formats, _ := atomicFormats.Load().([]format)
	atomicFormats.Store(append(formats, format{magic, decodeMeta}))
}

func DecodeMeta(r io.Reader) (Meta, error) {
	rr := asReader(r)
	formats, _ := atomicFormats.Load().([]format)

	for _, f := range formats {
		b, err := rr.Peek(len(f.magic))
		if err == nil && matchMagic(f.magic, b) {
			return f.decodeMeta(rr)
		}
	}

	if ok, err := IsSVG(rr); err != nil {
		return nil, err
	} else if ok {
		return &meta{format: "svg", width: 1, height: 1}, nil
	}

	return nil, ErrFormat
}
