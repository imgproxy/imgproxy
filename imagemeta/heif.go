package imagemeta

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/imgproxy/imgproxy/v3/imagetype"
)

const heifBoxHeaderSize = uint64(8)

var heicBrand = []byte("heic")
var avifBrand = []byte("avif")
var heifPict = []byte("pict")

type heifDiscarder interface {
	Discard(n int) (discarded int, err error)
}

type heifData struct {
	Format        imagetype.Type
	Width, Height int64
}

func (d *heifData) IsFilled() bool {
	return d.Format != imagetype.Unknown && d.Width > 0 && d.Height > 0
}

func heifReadN(r io.Reader, n uint64) (b []byte, err error) {
	if buf, ok := r.(*bytes.Buffer); ok {
		b = buf.Next(int(n))
		if len(b) == 0 {
			err = io.EOF
		}
		return
	}

	b = make([]byte, n)
	_, err = io.ReadFull(r, b)
	return
}

func heifDiscardN(r io.Reader, n uint64) error {
	if buf, ok := r.(*bytes.Buffer); ok {
		_ = buf.Next(int(n))
		return nil
	}

	if rd, ok := r.(heifDiscarder); ok {
		_, err := rd.Discard(int(n))
		return err
	}

	_, err := io.CopyN(io.Discard, r, int64(n))
	return err
}

func heifReadBoxHeader(r io.Reader) (boxType string, boxDataSize uint64, err error) {
	var b []byte

	b, err = heifReadN(r, heifBoxHeaderSize)
	if err != nil {
		return
	}

	headerSize := heifBoxHeaderSize

	boxDataSize = uint64(binary.BigEndian.Uint32(b[0:4]))
	boxType = string(b[4:8])

	if boxDataSize == 1 {
		b, err = heifReadN(r, 8)
		if err != nil {
			return
		}

		boxDataSize = (uint64(binary.BigEndian.Uint32(b[0:4])) << 32) |
			uint64(binary.BigEndian.Uint32(b[4:8]))
		headerSize += 8
	}

	if boxDataSize < heifBoxHeaderSize || boxDataSize > math.MaxInt64 {
		return "", 0, errors.New("Invalid box data size")
	}

	boxDataSize -= headerSize

	return
}

func heifAssignFormat(d *heifData, brand []byte) bool {
	if bytes.Equal(brand, heicBrand) {
		d.Format = imagetype.HEIC
		return true
	}

	if bytes.Equal(brand, avifBrand) {
		d.Format = imagetype.AVIF
		return true
	}

	return false
}

func heifReadFtyp(d *heifData, r io.Reader, boxDataSize uint64) error {
	if boxDataSize < 8 {
		return errors.New("Invalid ftyp data")
	}

	data, err := heifReadN(r, boxDataSize)
	if err != nil {
		return err
	}

	if heifAssignFormat(d, data[0:4]) {
		return nil
	}

	if boxDataSize >= 12 {
		for i := uint64(8); i < boxDataSize; i += 4 {
			if heifAssignFormat(d, data[i:i+4]) {
				return nil
			}
		}
	}

	return errors.New("Image is not compatible with heic/avif")
}

func heifReadMeta(d *heifData, r io.Reader, boxDataSize uint64) error {
	if boxDataSize < 4 {
		return errors.New("Invalid meta data")
	}

	data, err := heifReadN(r, boxDataSize)
	if err != nil {
		return err
	}

	if boxDataSize > 4 {
		if err := heifReadBoxes(d, bytes.NewBuffer(data[4:])); err != nil && err != io.EOF {
			return err
		}
	}

	return nil
}

func heifReadHldr(r io.Reader, boxDataSize uint64) error {
	if boxDataSize < 12 {
		return errors.New("Invalid hdlr data")
	}

	data, err := heifReadN(r, boxDataSize)
	if err != nil {
		return err
	}

	if !bytes.Equal(data[8:12], heifPict) {
		return fmt.Errorf("Invalid handler. Expected: pict, actual: %s", data[8:12])
	}

	return nil
}

func heifReadIspe(r io.Reader, boxDataSize uint64) (w, h int64, err error) {
	if boxDataSize < 12 {
		return 0, 0, errors.New("Invalid ispe data")
	}

	data, err := heifReadN(r, boxDataSize)
	if err != nil {
		return 0, 0, err
	}

	w = int64(binary.BigEndian.Uint32(data[4:8]))
	h = int64(binary.BigEndian.Uint32(data[8:12]))

	return
}

func heifReadBoxes(d *heifData, r io.Reader) error {
	for {
		boxType, boxDataSize, err := heifReadBoxHeader(r)
		if err != nil {
			return err
		}

		switch boxType {
		case "ftyp":
			if err := heifReadFtyp(d, r, boxDataSize); err != nil {
				return err
			}
		case "meta":
			if err := heifReadMeta(d, r, boxDataSize); err != nil {
				return err
			}
			if !d.IsFilled() {
				return errors.New("Dimensions data wasn't found in meta box")
			}
			return nil
		case "hdlr":
			if err := heifReadHldr(r, boxDataSize); err != nil {
				return nil
			}
		case "iprp", "ipco":
			data, err := heifReadN(r, boxDataSize)
			if err != nil {
				return err
			}

			if err := heifReadBoxes(d, bytes.NewBuffer(data)); err != nil && err != io.EOF {
				return err
			}
		case "ispe":
			w, h, err := heifReadIspe(r, boxDataSize)
			if err != nil {
				return err
			}
			if w > d.Width || h > d.Height {
				d.Width, d.Height = w, h
			}
		default:
			if err := heifDiscardN(r, boxDataSize); err != nil {
				return err
			}
		}
	}
}

func DecodeHeifMeta(r io.Reader) (Meta, error) {
	d := new(heifData)

	if err := heifReadBoxes(d, r); err != nil && !d.IsFilled() {
		return nil, err
	}

	return &meta{
		format: d.Format,
		width:  int(d.Width),
		height: int(d.Height),
	}, nil
}

func init() {
	RegisterFormat("????ftypheic", DecodeHeifMeta)
	RegisterFormat("????ftypheix", DecodeHeifMeta)
	RegisterFormat("????ftyphevc", DecodeHeifMeta)
	RegisterFormat("????ftypheim", DecodeHeifMeta)
	RegisterFormat("????ftypheis", DecodeHeifMeta)
	RegisterFormat("????ftyphevm", DecodeHeifMeta)
	RegisterFormat("????ftyphevs", DecodeHeifMeta)
	RegisterFormat("????ftypmif1", DecodeHeifMeta)
	RegisterFormat("????ftypavif", DecodeHeifMeta)
}
