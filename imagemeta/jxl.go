package imagemeta

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/imgproxy/imgproxy/v3/imagetype"
)

const (
	jxlCodestreamHeaderMinSize = 4
	jxlCodestreamHeaderMaxSize = 11
)

var jxlCodestreamMarker = []byte{0xff, 0x0a}
var jxlISOBMFFMarker = []byte{0x00, 0x00, 0x00, 0x0C, 0x4A, 0x58, 0x4C, 0x20, 0x0D, 0x0A, 0x87, 0x0A}

var jxlSizeSizes = []uint64{9, 13, 18, 30}

var jxlRatios = [][]uint64{
	{1, 1},
	{12, 10},
	{4, 3},
	{3, 2},
	{16, 9},
	{5, 4},
	{2, 1},
}

type jxlBitReader struct {
	buf    uint64
	bufLen uint64
}

func NewJxlBitReader(data []byte) *jxlBitReader {
	return &jxlBitReader{
		buf:    binary.LittleEndian.Uint64(data),
		bufLen: uint64(len(data) * 8),
	}
}

func (br *jxlBitReader) Read(n uint64) (uint64, error) {
	if n > br.bufLen {
		return 0, io.EOF
	}

	mask := uint64(1<<n) - 1
	res := br.buf & mask

	br.buf >>= n
	br.bufLen -= n

	return res, nil
}

func jxlReadJxlc(r io.Reader, boxDataSize uint64) ([]byte, error) {
	if boxDataSize < jxlCodestreamHeaderMinSize {
		return nil, newFormatError("JPEG XL", "invalid codestream box")
	}

	toRead := boxDataSize
	if toRead > jxlCodestreamHeaderMaxSize {
		toRead = jxlCodestreamHeaderMaxSize
	}

	return heifReadN(r, toRead)
}

func jxlReadJxlp(r io.Reader, boxDataSize uint64, codestream []byte) ([]byte, bool, error) {
	if boxDataSize < 4 {
		return nil, false, newFormatError("JPEG XL", "invalid jxlp box")
	}

	jxlpInd, err := heifReadN(r, 4)
	if err != nil {
		return nil, false, err
	}

	last := jxlpInd[0] == 0x80

	readLeft := jxlCodestreamHeaderMaxSize - len(codestream)
	if readLeft <= 0 {
		return codestream, last, nil
	}

	toRead := boxDataSize - 4
	if uint64(readLeft) < toRead {
		toRead = uint64(readLeft)
	}

	data, err := heifReadN(r, toRead)
	if err != nil {
		return nil, last, err
	}

	if codestream == nil {
		codestream = make([]byte, 0, jxlCodestreamHeaderMaxSize)
	}

	return append(codestream, data...), last, nil
}

// We can reuse HEIF functions to read ISO BMFF boxes
func jxlFindCodestream(r io.Reader) ([]byte, error) {
	var (
		codestream []byte
		last       bool
	)

	for {
		boxType, boxDataSize, err := heifReadBoxHeader(r)
		if err != nil {
			return nil, err
		}

		switch boxType {
		// jxlc box contins full codestream.
		// We can just read and return its header
		case "jxlc":
			codestream, err = jxlReadJxlc(r, boxDataSize)
			return codestream, err

		// jxlp partial codestream.
		// We should read its data until we read jxlCodestreamHeaderSize bytes
		case "jxlp":
			codestream, last, err = jxlReadJxlp(r, boxDataSize, codestream)
			if err != nil {
				return nil, err
			}

			csLen := len(codestream)
			if csLen >= jxlCodestreamHeaderMaxSize || (last && csLen >= jxlCodestreamHeaderMinSize) {
				return codestream, nil
			}

			if last {
				return nil, newFormatError("JPEG XL", "invalid codestream box")
			}

		// Skip other boxes
		default:
			if err := heifDiscardN(r, boxDataSize); err != nil {
				return nil, err
			}
		}
	}
}

func jxlParseSize(br *jxlBitReader, small bool) (uint64, error) {
	if small {
		size, err := br.Read(5)
		return (size + 1) * 8, err
	} else {
		selector, err := br.Read(2)
		if err != nil {
			return 0, err
		}

		sizeSize := jxlSizeSizes[selector]
		size, err := br.Read(sizeSize)

		return size + 1, err
	}
}

func jxlDecodeCodestreamHeader(buf []byte) (width, height uint64, err error) {
	if len(buf) < jxlCodestreamHeaderMinSize {
		return 0, 0, newFormatError("JPEG XL", "invalid codestream header")
	}

	if !bytes.Equal(buf[0:2], jxlCodestreamMarker) {
		return 0, 0, newFormatError("JPEG XL", "missing codestream marker")
	}

	br := NewJxlBitReader(buf[2:])

	smallBit, sbErr := br.Read(1)
	if sbErr != nil {
		return 0, 0, sbErr
	}

	small := smallBit == 1

	height, err = jxlParseSize(br, small)
	if err != nil {
		return 0, 0, err
	}

	ratioIdx, riErr := br.Read(3)
	if riErr != nil {
		return 0, 0, riErr
	}

	if ratioIdx == 0 {
		width, err = jxlParseSize(br, small)
	} else {
		ratio := jxlRatios[ratioIdx-1]
		width = height * ratio[0] / ratio[1]
	}

	return
}

func DecodeJxlMeta(r io.Reader) (Meta, error) {
	var (
		tmp           [12]byte
		codestream    []byte
		width, height uint64
		err           error
	)

	if _, err = io.ReadFull(r, tmp[:2]); err != nil {
		return nil, err
	}

	if bytes.Equal(tmp[0:2], jxlCodestreamMarker) {
		if _, err = io.ReadFull(r, tmp[2:]); err != nil {
			return nil, err
		}

		codestream = tmp[:]
	} else {
		if _, err = io.ReadFull(r, tmp[2:12]); err != nil {
			return nil, err
		}

		if !bytes.Equal(tmp[0:12], jxlISOBMFFMarker) {
			return nil, newFormatError("JPEG XL", "invalid header")
		}

		codestream, err = jxlFindCodestream(r)
		if err != nil {
			return nil, err
		}
	}

	width, height, err = jxlDecodeCodestreamHeader(codestream)
	if err != nil {
		return nil, err
	}

	return &meta{
		format: imagetype.JXL,
		width:  int(width),
		height: int(height),
	}, nil
}

func init() {
	RegisterFormat(string(jxlCodestreamMarker), DecodeJxlMeta)
	RegisterFormat(string(jxlISOBMFFMarker), DecodeJxlMeta)
}
