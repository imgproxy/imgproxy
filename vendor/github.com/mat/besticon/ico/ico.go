// Package ico registers image.Decode and DecodeConfig support
// for the icon (container) format.
package ico

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"io"
	"io/ioutil"

	"image/png"

	"golang.org/x/image/bmp"
)

type icondir struct {
	Reserved uint16
	Type     uint16
	Count    uint16
	Entries  []icondirEntry
}

type icondirEntry struct {
	Width        byte
	Height       byte
	PaletteCount byte
	Reserved     byte
	ColorPlanes  uint16
	BitsPerPixel uint16
	Size         uint32
	Offset       uint32
}

func (dir *icondir) FindBestIcon() *icondirEntry {
	if len(dir.Entries) == 0 {
		return nil
	}

	best := dir.Entries[0]
	for _, e := range dir.Entries {
		if (e.width() > best.width()) && (e.height() > best.height()) {
			best = e
		}
	}
	return &best
}

// ParseIco parses the icon and returns meta information for the icons as icondir.
func ParseIco(r io.Reader) (*icondir, error) {
	dir := icondir{}

	var err error
	err = binary.Read(r, binary.LittleEndian, &dir.Reserved)
	if err != nil {
		return nil, err
	}

	err = binary.Read(r, binary.LittleEndian, &dir.Type)
	if err != nil {
		return nil, err
	}

	err = binary.Read(r, binary.LittleEndian, &dir.Count)
	if err != nil {
		return nil, err
	}

	for i := uint16(0); i < dir.Count; i++ {
		entry := icondirEntry{}
		e := parseIcondirEntry(r, &entry)
		if e != nil {
			return nil, e
		}
		dir.Entries = append(dir.Entries, entry)
	}

	return &dir, err
}

func parseIcondirEntry(r io.Reader, e *icondirEntry) error {
	err := binary.Read(r, binary.LittleEndian, e)
	if err != nil {
		return err
	}

	return nil
}

type dibHeader struct {
	dibHeaderSize uint32
	width         uint32
	height        uint32
}

func (e *icondirEntry) ColorCount() int {
	if e.PaletteCount == 0 {
		return 256
	}
	return int(e.PaletteCount)
}

func (e *icondirEntry) width() int {
	if e.Width == 0 {
		return 256
	}
	return int(e.Width)
}

func (e *icondirEntry) height() int {
	if e.Height == 0 {
		return 256
	}
	return int(e.Height)
}

// DecodeConfig returns just the dimensions of the largest image
// contained in the icon withou decoding the entire icon file.
func DecodeConfig(r io.Reader) (image.Config, error) {
	dir, err := ParseIco(r)
	if err != nil {
		return image.Config{}, err
	}

	best := dir.FindBestIcon()
	if best == nil {
		return image.Config{}, errInvalid
	}
	return image.Config{Width: best.width(), Height: best.height()}, nil
}

// The bitmap header structure we read from an icondirEntry
type bitmapHeaderRead struct {
	Size            uint32
	Width           uint32
	Height          uint32
	Planes          uint16
	BitCount        uint16
	Compression     uint32
	ImageSize       uint32
	XPixelsPerMeter uint32
	YPixelsPerMeter uint32
	ColorsUsed      uint32
	ColorsImportant uint32
}

// The bitmap header structure we need to generate for bmp.Decode()
type bitmapHeaderWrite struct {
	sigBM           [2]byte
	fileSize        uint32
	resverved       [2]uint16
	pixOffset       uint32
	Size            uint32
	Width           uint32
	Height          uint32
	Planes          uint16
	BitCount        uint16
	Compression     uint32
	ImageSize       uint32
	XPixelsPerMeter uint32
	YPixelsPerMeter uint32
	ColorsUsed      uint32
	ColorsImportant uint32
}

var errInvalid = errors.New("ico: invalid ICO image")

// Decode returns the largest image contained in the icon
// which might be a bmp or png
func Decode(r io.Reader) (image.Image, error) {
	icoBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	r = bytes.NewReader(icoBytes)
	dir, err := ParseIco(r)
	if err != nil {
		return nil, errInvalid
	}

	best := dir.FindBestIcon()
	if best == nil {
		return nil, errInvalid
	}

	return parseImage(best, icoBytes)
}

func parseImage(entry *icondirEntry, icoBytes []byte) (image.Image, error) {
	r := bytes.NewReader(icoBytes)
	r.Seek(int64(entry.Offset), 0)

	// Try PNG first then BMP
	img, err := png.Decode(r)
	if err != nil {
		return parseBMP(entry, icoBytes)
	}
	return img, nil
}

func parseBMP(entry *icondirEntry, icoBytes []byte) (image.Image, error) {
	bmpBytes, err := makeFullBMPBytes(entry, icoBytes)
	if err != nil {
		return nil, err
	}
	return bmp.Decode(bmpBytes)
}

func makeFullBMPBytes(entry *icondirEntry, icoBytes []byte) (*bytes.Buffer, error) {
	r := bytes.NewReader(icoBytes)
	r.Seek(int64(entry.Offset), 0)

	var err error
	h := bitmapHeaderRead{}

	err = binary.Read(r, binary.LittleEndian, &h)
	if err != nil {
		return nil, err
	}

	if h.Size != 40 || h.Planes != 1 {
		return nil, errInvalid
	}

	var pixOffset uint32
	if h.ColorsUsed == 0 && h.BitCount <= 8 {
		pixOffset = 14 + 40 + 4*(1<<h.BitCount)
	} else {
		pixOffset = 14 + 40 + 4*h.ColorsUsed
	}

	writeHeader := &bitmapHeaderWrite{
		sigBM:           [2]byte{'B', 'M'},
		fileSize:        14 + 40 + uint32(len(icoBytes)), // correct? important?
		pixOffset:       pixOffset,
		Size:            40,
		Width:           uint32(h.Width),
		Height:          uint32(h.Height / 2),
		Planes:          h.Planes,
		BitCount:        h.BitCount,
		Compression:     h.Compression,
		ColorsUsed:      h.ColorsUsed,
		ColorsImportant: h.ColorsImportant,
	}

	buf := new(bytes.Buffer)
	if err = binary.Write(buf, binary.LittleEndian, writeHeader); err != nil {
		return nil, err
	}
	io.CopyN(buf, r, int64(entry.Size))

	return buf, nil
}

const icoHeader = "\x00\x00\x01\x00"

func init() {
	image.RegisterFormat("ico", icoHeader, Decode, DecodeConfig)
}
