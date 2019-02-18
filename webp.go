// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Original code was cropped and fixed by @DarthSim for imgproxy needs

package main

import (
	"errors"
	"image"
	"image/color"
	"io"

	"golang.org/x/image/riff"
	"golang.org/x/image/vp8"
	"golang.org/x/image/vp8l"
)

var errInvalidFormat = errors.New("webp: invalid format")

var (
	fccALPH = riff.FourCC{'A', 'L', 'P', 'H'}
	fccVP8  = riff.FourCC{'V', 'P', '8', ' '}
	fccVP8L = riff.FourCC{'V', 'P', '8', 'L'}
	fccVP8X = riff.FourCC{'V', 'P', '8', 'X'}
	fccWEBP = riff.FourCC{'W', 'E', 'B', 'P'}
)

// Since we need this only for type detecting, we can return fake image
func decodeWebp(r io.Reader) (image.Image, error) {
	return image.NewRGBA(image.Rect(0, 0, 1, 1)), nil
}

func decodeWebpConfig(r io.Reader) (image.Config, error) {
	formType, riffReader, err := riff.NewReader(r)
	if err != nil {
		return image.Config{}, err
	}
	if formType != fccWEBP {
		return image.Config{}, errInvalidFormat
	}

	var (
		wantAlpha bool
		buf       [10]byte
	)

	for {
		chunkID, chunkLen, chunkData, err := riffReader.Next()
		if err == io.EOF {
			err = errInvalidFormat
		}
		if err != nil {
			return image.Config{}, err
		}

		switch chunkID {
		case fccALPH:
			// Ignore
		case fccVP8:
			if wantAlpha || int32(chunkLen) < 0 {
				return image.Config{}, errInvalidFormat
			}

			d := vp8.NewDecoder()
			d.Init(chunkData, int(chunkLen))

			fh, err := d.DecodeFrameHeader()

			return image.Config{
				ColorModel: color.YCbCrModel,
				Width:      fh.Width,
				Height:     fh.Height,
			}, err

		case fccVP8L:
			if wantAlpha {
				return image.Config{}, errInvalidFormat
			}
			return vp8l.DecodeConfig(chunkData)

		case fccVP8X:
			if chunkLen != 10 {
				return image.Config{}, errInvalidFormat
			}

			if _, err := io.ReadFull(chunkData, buf[:10]); err != nil {
				return image.Config{}, err
			}

			const alphaBit = 1 << 4

			wantAlpha = buf[0] != alphaBit
			widthMinusOne := uint32(buf[4]) | uint32(buf[5])<<8 | uint32(buf[6])<<16
			heightMinusOne := uint32(buf[7]) | uint32(buf[8])<<8 | uint32(buf[9])<<16

			return image.Config{
				ColorModel: color.NYCbCrAModel,
				Width:      int(widthMinusOne) + 1,
				Height:     int(heightMinusOne) + 1,
			}, nil

		default:
			return image.Config{}, errInvalidFormat
		}
	}
}

func init() {
	image.RegisterFormat("webp", "RIFF????WEBPVP8", decodeWebp, decodeWebpConfig)
}
