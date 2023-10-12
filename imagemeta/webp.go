// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Original code was cropped and fixed by @DarthSim for imgproxy needs

package imagemeta

import (
	"errors"
	"io"

	"github.com/imgproxy/imgproxy/v3/imagetype"
	"golang.org/x/image/riff"
	"golang.org/x/image/vp8"
	"golang.org/x/image/vp8l"
)

var ErrWebpInvalidFormat = errors.New("webp: invalid format")

var (
	webpFccALPH = riff.FourCC{'A', 'L', 'P', 'H'}
	webpFccVP8  = riff.FourCC{'V', 'P', '8', ' '}
	webpFccVP8L = riff.FourCC{'V', 'P', '8', 'L'}
	webpFccVP8X = riff.FourCC{'V', 'P', '8', 'X'}
	webpFccWEBP = riff.FourCC{'W', 'E', 'B', 'P'}
)

func DecodeWebpMeta(r io.Reader) (Meta, error) {
	formType, riffReader, err := riff.NewReader(r)
	if err != nil {
		return nil, err
	}
	if formType != webpFccWEBP {
		return nil, ErrWebpInvalidFormat
	}

	var buf [10]byte

	for {
		chunkID, chunkLen, chunkData, err := riffReader.Next()
		if err == io.EOF {
			err = ErrWebpInvalidFormat
		}
		if err != nil {
			return nil, err
		}

		switch chunkID {
		case webpFccALPH:
			// Ignore
		case webpFccVP8:
			if int32(chunkLen) < 0 {
				return nil, ErrWebpInvalidFormat
			}

			d := vp8.NewDecoder()
			d.Init(chunkData, int(chunkLen))

			fh, err := d.DecodeFrameHeader()

			return &meta{
				format: imagetype.WEBP,
				width:  fh.Width,
				height: fh.Height,
			}, err

		case webpFccVP8L:
			conf, err := vp8l.DecodeConfig(chunkData)
			if err != nil {
				return nil, err
			}

			return &meta{
				format: imagetype.WEBP,
				width:  conf.Width,
				height: conf.Height,
			}, nil

		case webpFccVP8X:
			if chunkLen != 10 {
				return nil, ErrWebpInvalidFormat
			}

			if _, err := io.ReadFull(chunkData, buf[:10]); err != nil {
				return nil, err
			}

			widthMinusOne := uint32(buf[4]) | uint32(buf[5])<<8 | uint32(buf[6])<<16
			heightMinusOne := uint32(buf[7]) | uint32(buf[8])<<8 | uint32(buf[9])<<16

			return &meta{
				format: imagetype.WEBP,
				width:  int(widthMinusOne) + 1,
				height: int(heightMinusOne) + 1,
			}, nil

		default:
			return nil, ErrWebpInvalidFormat
		}
	}
}

func init() {
	RegisterFormat("RIFF????WEBPVP8", DecodeWebpMeta)
}
