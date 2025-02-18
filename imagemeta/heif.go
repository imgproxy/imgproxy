package imagemeta

import (
	"bytes"
	"cmp"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"slices"

	"github.com/imgproxy/imgproxy/v3/imagetype"
)

const heifBoxHeaderSize = uint64(8)

var heicBrand = []byte("heic")
var heixBrand = []byte("heix")
var avifBrand = []byte("avif")
var heifPict = []byte("pict")

type heifDiscarder interface {
	Discard(n int) (discarded int, err error)
}

type heifSize struct {
	Width, Height int64
}

type heifData struct {
	Format imagetype.Type
	Sizes  []heifSize
}

func (d *heifData) Meta() (*meta, error) {
	if d.Format == imagetype.Unknown {
		return nil, newFormatError("HEIF", "format data wasn't found")
	}

	if len(d.Sizes) == 0 {
		return nil, newFormatError("HEIF", "dimensions data wasn't found")
	}

	bestSize := slices.MaxFunc(d.Sizes, func(a, b heifSize) int {
		return cmp.Compare(a.Width*a.Height, b.Width*b.Height)
	})

	return &meta{
		format: d.Format,
		width:  int(bestSize.Width),
		height: int(bestSize.Height),
	}, nil
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
		return "", 0, newFormatError("HEIF", "invalid box data size")
	}

	boxDataSize -= headerSize

	return
}

func heifAssignFormat(d *heifData, brand []byte) bool {
	if bytes.Equal(brand, heicBrand) || bytes.Equal(brand, heixBrand) {
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
		return newFormatError("HEIF", "invalid ftyp data")
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

	return newFormatError("HEIF", "image is not compatible with heic/avif")
}

func heifReadMeta(d *heifData, r io.Reader, boxDataSize uint64) error {
	if boxDataSize < 4 {
		return newFormatError("HEIF", "invalid meta data")
	}

	data, err := heifReadN(r, boxDataSize)
	if err != nil {
		return err
	}

	if boxDataSize > 4 {
		if err := heifReadBoxes(d, bytes.NewBuffer(data[4:])); err != nil && !errors.Is(err, io.EOF) {
			return err
		}
	}

	return nil
}

func heifReadHldr(r io.Reader, boxDataSize uint64) error {
	if boxDataSize < 12 {
		return newFormatError("HEIF", "invalid hdlr data")
	}

	data, err := heifReadN(r, boxDataSize)
	if err != nil {
		return err
	}

	if !bytes.Equal(data[8:12], heifPict) {
		return newFormatError("HEIF", fmt.Sprintf("Invalid handler. Expected: pict, actual: %s", data[8:12]))
	}

	return nil
}

func heifReadIspe(r io.Reader, boxDataSize uint64) (w, h int64, err error) {
	if boxDataSize < 12 {
		return 0, 0, newFormatError("HEIF", "invalid ispe data")
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
			return heifReadMeta(d, r, boxDataSize)
		case "hdlr":
			if err := heifReadHldr(r, boxDataSize); err != nil {
				return err
			}
		case "iprp", "ipco":
			data, err := heifReadN(r, boxDataSize)
			if err != nil {
				return err
			}

			if err := heifReadBoxes(d, bytes.NewBuffer(data)); err != nil && !errors.Is(err, io.EOF) {
				return err
			}
		case "ispe":
			w, h, err := heifReadIspe(r, boxDataSize)
			if err != nil {
				return err
			}
			d.Sizes = append(d.Sizes, heifSize{Width: w, Height: h})
		case "irot":
			data, err := heifReadN(r, boxDataSize)
			if err != nil {
				return err
			}
			if len(d.Sizes) > 0 && len(data) > 0 && (data[0] == 1 || data[0] == 3) {
				lastSize := d.Sizes[len(d.Sizes)-1]
				d.Sizes[len(d.Sizes)-1] = heifSize{Width: lastSize.Height, Height: lastSize.Width}
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

	if err := heifReadBoxes(d, r); err != nil {
		return nil, err
	}

	return d.Meta()
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
