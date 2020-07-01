package imagemeta

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const heicBoxHeaderSize = int64(8)

var heicBrand = []byte("heic")
var heicPict = []byte("pict")

type heicDimensionsData struct {
	Width, Height int64
}

func (d *heicDimensionsData) IsFilled() bool {
	return d.Width > 0 && d.Height > 0
}

func heicReadBoxHeader(r io.Reader) (boxType string, boxDataSize int64, err error) {
	b := make([]byte, heicBoxHeaderSize)
	_, err = r.Read(b)
	if err != nil {
		return
	}

	boxDataSize = int64(binary.BigEndian.Uint32(b[0:4])) - heicBoxHeaderSize
	boxType = string(b[4:8])

	return
}

func heicReadBoxData(r io.Reader, boxDataSize int64) (b []byte, err error) {
	b = make([]byte, boxDataSize)
	_, err = r.Read(b)
	return
}

func heicReadFtyp(r io.Reader, boxDataSize int64) error {
	if boxDataSize < 8 {
		return errors.New("Invalid ftyp data")
	}

	data, err := heicReadBoxData(r, boxDataSize)
	if err != nil {
		return err
	}

	if bytes.Equal(data[0:4], heicBrand) {
		return nil
	}

	if boxDataSize >= 12 {
		for i := int64(8); i < boxDataSize; i += 4 {
			if bytes.Equal(data[i:i+4], heicBrand) {
				return nil
			}
		}
	}

	return errors.New("Image is not compatible with heic")
}

func heicReadMeta(d *heicDimensionsData, r io.Reader, boxDataSize int64) error {
	if boxDataSize < 4 {
		return errors.New("Invalid meta data")
	}

	if _, err := r.Read(make([]byte, 4)); err != nil {
		return err
	}

	if boxDataSize > 4 {
		if err := heicReadBoxes(d, io.LimitReader(r, boxDataSize-4)); err != nil && err != io.EOF {
			return err
		}
	}

	return nil
}

func heicReadHldr(r io.Reader, boxDataSize int64) error {
	if boxDataSize < 12 {
		return errors.New("Invalid hdlr data")
	}

	data, err := heicReadBoxData(r, boxDataSize)
	if err != nil {
		return err
	}

	if !bytes.Equal(data[8:12], heicPict) {
		return fmt.Errorf("Invalid handler. Expected: pict, actual: %s", data[8:12])
	}

	return nil
}

func heicReadIspe(r io.Reader, boxDataSize int64) (w, h int64, err error) {
	if boxDataSize < 12 {
		return 0, 0, errors.New("Invalid ispe data")
	}

	data, err := heicReadBoxData(r, boxDataSize)
	if err != nil {
		return 0, 0, err
	}

	w = int64(binary.BigEndian.Uint32(data[4:8]))
	h = int64(binary.BigEndian.Uint32(data[8:12]))

	return
}

func heicReadBoxes(d *heicDimensionsData, r io.Reader) error {
	for {
		boxType, boxDataSize, err := heicReadBoxHeader(r)

		if err != nil {
			return err
		}

		if boxDataSize < 0 {
			return errors.New("Invalid box data")
		}

		// log.Printf("Box type: %s; Box data size: %d", boxType, boxDataSize)

		switch boxType {
		case "ftyp":
			if err := heicReadFtyp(r, boxDataSize); err != nil {
				return err
			}
		case "meta":
			if err := heicReadMeta(d, r, boxDataSize); err != nil {
				return err
			}
			if !d.IsFilled() {
				return errors.New("Dimensions data wasn't found in meta box")
			}
			return nil
		case "hdlr":
			if err := heicReadHldr(r, boxDataSize); err != nil {
				return nil
			}
		case "iprp", "ipco":
			if err := heicReadBoxes(d, io.LimitReader(r, boxDataSize)); err != nil && err != io.EOF {
				return err
			}
		case "ispe":
			w, h, err := heicReadIspe(r, boxDataSize)
			if err != nil {
				return err
			}
			if w > d.Width || h > d.Height {
				d.Width, d.Height = w, h
			}
		case "mdat":
			return errors.New("mdat box occurred before meta box")
		default:
			if _, err := heicReadBoxData(r, boxDataSize); err != nil {
				return err
			}
		}
	}
}

func DecodeHeicMeta(r io.Reader) (Meta, error) {
	d := new(heicDimensionsData)

	if err := heicReadBoxes(d, r); err != nil && !d.IsFilled() {
		return nil, err
	}

	return &meta{
		format: "heic",
		width:  int(d.Width),
		height: int(d.Height),
	}, nil
}

func init() {
	RegisterFormat("????ftypheic", DecodeHeicMeta)
	RegisterFormat("????ftypheix", DecodeHeicMeta)
	RegisterFormat("????ftyphevc", DecodeHeicMeta)
	RegisterFormat("????ftypheim", DecodeHeicMeta)
	RegisterFormat("????ftypheis", DecodeHeicMeta)
	RegisterFormat("????ftyphevm", DecodeHeicMeta)
	RegisterFormat("????ftyphevs", DecodeHeicMeta)
	RegisterFormat("????ftypmif1", DecodeHeicMeta)
}
