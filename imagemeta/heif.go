package imagemeta

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const heifBoxHeaderSize = int64(8)

var heicBrand = []byte("heic")
var avifBrand = []byte("avif")
var heifPict = []byte("pict")

type heifData struct {
	Format        string
	Width, Height int64
}

func (d *heifData) IsFilled() bool {
	return len(d.Format) > 0 && d.Width > 0 && d.Height > 0
}

func heifReadBoxHeader(r io.Reader) (boxType string, boxDataSize int64, err error) {
	b := make([]byte, heifBoxHeaderSize)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return
	}

	boxDataSize = int64(binary.BigEndian.Uint32(b[0:4])) - heifBoxHeaderSize
	boxType = string(b[4:8])

	return
}

func heifReadBoxData(r io.Reader, boxDataSize int64) (b []byte, err error) {
	b = make([]byte, boxDataSize)
	_, err = io.ReadFull(r, b)
	return
}

func heifAssignFormat(d *heifData, brand []byte) bool {
	if bytes.Equal(brand, heicBrand) {
		d.Format = "heic"
		return true
	}

	if bytes.Equal(brand, avifBrand) {
		d.Format = "avif"
		return true
	}

	return false
}

func heifReadFtyp(d *heifData, r io.Reader, boxDataSize int64) error {
	if boxDataSize < 8 {
		return errors.New("Invalid ftyp data")
	}

	data, err := heifReadBoxData(r, boxDataSize)
	if err != nil {
		return err
	}

	if heifAssignFormat(d, data[0:4]) {
		return nil
	}

	if boxDataSize >= 12 {
		for i := int64(8); i < boxDataSize; i += 4 {
			if heifAssignFormat(d, data[i:i+4]) {
				return nil
			}
		}
	}

	return errors.New("Image is not compatible with heic/avif")
}

func heifReadMeta(d *heifData, r io.Reader, boxDataSize int64) error {
	if boxDataSize < 4 {
		return errors.New("Invalid meta data")
	}

	if _, err := io.ReadFull(r, make([]byte, 4)); err != nil {
		return err
	}

	if boxDataSize > 4 {
		if err := heifReadBoxes(d, io.LimitReader(r, boxDataSize-4)); err != nil && err != io.EOF {
			return err
		}
	}

	return nil
}

func heifReadHldr(r io.Reader, boxDataSize int64) error {
	if boxDataSize < 12 {
		return errors.New("Invalid hdlr data")
	}

	data, err := heifReadBoxData(r, boxDataSize)
	if err != nil {
		return err
	}

	if !bytes.Equal(data[8:12], heifPict) {
		return fmt.Errorf("Invalid handler. Expected: pict, actual: %s", data[8:12])
	}

	return nil
}

func heifReadIspe(r io.Reader, boxDataSize int64) (w, h int64, err error) {
	if boxDataSize < 12 {
		return 0, 0, errors.New("Invalid ispe data")
	}

	data, err := heifReadBoxData(r, boxDataSize)
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

		if boxDataSize < 0 {
			return errors.New("Invalid box data")
		}

		// log.Printf("Box type: %s; Box data size: %d", boxType, boxDataSize)

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
			if err := heifReadBoxes(d, io.LimitReader(r, boxDataSize)); err != nil && err != io.EOF {
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
		case "mdat":
			return errors.New("mdat box occurred before meta box")
		default:
			if _, err := heifReadBoxData(r, boxDataSize); err != nil {
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
