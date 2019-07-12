package main

import (
	"encoding/binary"
	"image"
	"image/color"
	"io"
	"math"
)

const (
	leHeader = "II\x2A\x00" // Header for little-endian files.
	beHeader = "MM\x00\x2A" // Header for big-endian files.

	ifdLen = 12 // Length of an IFD entry in bytes.

	// Tags (see p. 28-41 of the spec).
	tImageWidth = 256
	tImageLength = 257
	
	// Data types (p. 14-16 of the spec).
	dtByte     = 1
	dtASCII    = 2
	dtShort    = 3
	dtLong     = 4
  dtRational = 5
)

// The length of one instance of each data type in bytes.
var lengths = [...]uint32{0, 1, 1, 2, 4, 8}

// imageMode represents the mode of the image.
type imageMode int

// A FormatError reports that the input is not a valid TIFF image.
type FormatError string

func (e FormatError) Error() string {
	return "tiff: invalid format: " + string(e)
}

// An UnsupportedError reports that the input uses a valid but
// unimplemented feature.
type UnsupportedError string

func (e UnsupportedError) Error() string {
	return "tiff: unsupported feature: " + string(e)
}

// buffer buffers an io.Reader to satisfy io.ReaderAt.
type buffer struct {
	r   io.Reader
	buf []byte
}

// fill reads data from b.r until the buffer contains at least end bytes.
func (b *buffer) fill(end int) error {
	m := len(b.buf)
	if end > m {
		if end > cap(b.buf) {
			newcap := 1024
			for newcap < end {
				newcap *= 2
			}
			newbuf := make([]byte, end, newcap)
			copy(newbuf, b.buf)
			b.buf = newbuf
		} else {
			b.buf = b.buf[:end]
		}
		if n, err := io.ReadFull(b.r, b.buf[m:end]); err != nil {
			end = m + n
			b.buf = b.buf[:end]
			return err
		}
	}
	return nil
}

func (b *buffer) ReadAt(p []byte, off int64) (int, error) {
	o := int(off)
	end := o + len(p)
	if int64(end) != off+int64(len(p)) {
		return 0, io.ErrUnexpectedEOF
	}

	err := b.fill(end)
	return copy(p, b.buf[o:end]), err
}

type decoder struct {
	r         io.ReaderAt
	byteOrder binary.ByteOrder
	config    image.Config
	mode      imageMode
	bpp       uint
	features  map[int][]uint
	palette   []color.Color

	buf   []byte
	off   int    // Current offset in buf.
	v     uint32 // Buffer value for reading with arbitrary bit depths.
	nbits uint   // Remaining number of bits in v.
}

// firstVal returns the first uint of the features entry with the given tag,
// or 0 if the tag does not exist.
func (d *decoder) firstVal(tag int) uint {
	f := d.features[tag]
	if len(f) == 0 {
		return 0
	}
	return f[0]
}

// ifdUint decodes the IFD entry in p, which must be of the Byte, Short
// or Long type, and returns the decoded uint values.
func (d *decoder) ifdUint(p []byte) (u []uint, err error) {
	var raw []byte
	if len(p) < ifdLen {
		return nil, FormatError("bad IFD entry")
	}

	datatype := d.byteOrder.Uint16(p[2:4])
	if dt := int(datatype); dt <= 0 || dt >= len(lengths) {
		return nil, UnsupportedError("IFD entry datatype")
	}

	count := d.byteOrder.Uint32(p[4:8])
	if count > math.MaxInt32/lengths[datatype] {
		return nil, FormatError("IFD data too large")
	}
	if datalen := lengths[datatype] * count; datalen > 4 {
		// The IFD contains a pointer to the real value.
		raw = make([]byte, datalen)
		_, err = d.r.ReadAt(raw, int64(d.byteOrder.Uint32(p[8:12])))
	} else {
		raw = p[8 : 8+datalen]
	}
	if err != nil {
		return nil, err
	}

	u = make([]uint, count)
	switch datatype {
	case dtByte:
		for i := uint32(0); i < count; i++ {
			u[i] = uint(raw[i])
		}
	case dtShort:
		for i := uint32(0); i < count; i++ {
			u[i] = uint(d.byteOrder.Uint16(raw[2*i : 2*(i+1)]))
		}
	case dtLong:
		for i := uint32(0); i < count; i++ {
			u[i] = uint(d.byteOrder.Uint32(raw[4*i : 4*(i+1)]))
		}
	default:
		return nil, UnsupportedError("data type")
	}
	return u, nil
}

// parseIFD decides whether the IFD entry in p is "interesting" and
// stows away the data in the decoder. It returns the tag number of the
// entry and an error, if any.
func (d *decoder) parseIFD(p []byte) (int, error) {
	tag := d.byteOrder.Uint16(p[0:2])
	switch tag {
	case tImageLength, tImageWidth:
		val, err := d.ifdUint(p)
		if err != nil {
			return 0, err
		}
		d.features[int(tag)] = val
	}
	return int(tag), nil
}

// newReaderAt converts an io.Reader into an io.ReaderAt.
func newReaderAt(r io.Reader) io.ReaderAt {
	if ra, ok := r.(io.ReaderAt); ok {
		return ra
	}
	return &buffer{
		r:   r,
		buf: make([]byte, 0, 1024),
	}
}

func newDecoder(r io.Reader) (*decoder, error) {
	d := &decoder{
		r:        newReaderAt(r),
		features: make(map[int][]uint),
	}

	p := make([]byte, 8)
	if _, err := d.r.ReadAt(p, 0); err != nil {
		return nil, err
	}
	switch string(p[0:4]) {
	case leHeader:
		d.byteOrder = binary.LittleEndian
	case beHeader:
		d.byteOrder = binary.BigEndian
	default:
		return nil, FormatError("malformed header")
	}

	ifdOffset := int64(d.byteOrder.Uint32(p[4:8]))

	// The first two bytes contain the number of entries (12 bytes each).
	if _, err := d.r.ReadAt(p[0:2], ifdOffset); err != nil {
		return nil, err
	}
	numItems := int(d.byteOrder.Uint16(p[0:2]))

	// All IFD entries are read in one chunk.
	p = make([]byte, ifdLen*numItems)
	if _, err := d.r.ReadAt(p, ifdOffset+2); err != nil {
		return nil, err
	}

	prevTag := -1
	for i := 0; i < len(p); i += ifdLen {
		tag, err := d.parseIFD(p[i : i+ifdLen])
		if err != nil {
			return nil, err
		}
		if tag <= prevTag {
			return nil, FormatError("tags are not sorted in ascending order")
		}
		prevTag = tag
	}

	d.config.Width = int(d.firstVal(tImageWidth))
	d.config.Height = int(d.firstVal(tImageLength))

	return d, nil
}

// DecodeConfig returns dimensions of a TIFF image without
// decoding the entire image.
func DecodeConfig(r io.Reader) (image.Config, error) {
	d, err := newDecoder(r)
	if err != nil {
		return image.Config{}, err
	}
	return d.config, nil
}

func init() {
	// Register fake tiff decoder. Since we need this only for type detecting, we can
	// return fake image sizes
	decode := func(io.Reader) (image.Image, error) {
		return image.NewRGBA(image.Rect(0, 0, 1, 1)), nil
	}
	// decodeConfig := func(io.Reader) (image.Config, error) {
	// 	return image.Config{ColorModel: color.RGBAModel, Width: 1, Height: 1}, nil
	// }

	image.RegisterFormat("tiff", leHeader, decode, DecodeConfig)
	image.RegisterFormat("tiff", beHeader, decode, DecodeConfig)
}
