package imagemeta

import (
	"bufio"
	"io"

	"github.com/imgproxy/imgproxy/v3/imagetype"
)

const (
	// https://www.disktuna.com/list-of-jpeg-markers/
	jpegSof0Marker  = 0xc0 // Start Of Frame (Baseline Sequential).
	jpegSof1Marker  = 0xc1 // Start Of Frame (Extended Sequential DCT)
	jpegSof2Marker  = 0xc2 // Start Of Frame (Progressive DCT )
	jpegSof3Marker  = 0xc3 // Start Of Frame (Lossless sequential)
	jpegSof5Marker  = 0xc5 // Start Of Frame (Differential sequential DCT)
	jpegSof6Marker  = 0xc6 // Start Of Frame (Differential progressive DCT)
	jpegSof7Marker  = 0xc7 // Start Of Frame (Differential lossless sequential)
	jpegSof9Marker  = 0xc9 // Start Of Frame (Extended sequential DCT, Arithmetic coding)
	jpegSof10Marker = 0xca // Start Of Frame (Progressive DCT, Arithmetic coding)
	jpegSof11Marker = 0xcb // Start Of Frame (Lossless sequential, Arithmetic coding)
	jpegSof13Marker = 0xcd // Start Of Frame (Differential sequential DCT, Arithmetic coding)
	jpegSof14Marker = 0xce // Start Of Frame (Differential progressive DCT, Arithmetic coding)
	jpegSof15Marker = 0xcf // Start Of Frame (Differential lossless sequential, Arithmetic coding).
	jpegRst0Marker  = 0xd0 // ReSTart (0).
	jpegRst7Marker  = 0xd7 // ReSTart (7).
	jpegSoiMarker   = 0xd8 // Start Of Image.
	jpegEoiMarker   = 0xd9 // End Of Image.
	jpegSosMarker   = 0xda // Start Of Scan.
)

type jpegReader interface {
	io.Reader
	ReadByte() (byte, error)
	Discard(n int) (discarded int, err error)
}

func asJpegReader(r io.Reader) jpegReader {
	if rr, ok := r.(jpegReader); ok {
		return rr
	}
	return bufio.NewReader(r)
}

type JpegFormatError string

func (e JpegFormatError) Error() string { return "invalid JPEG format: " + string(e) }

func DecodeJpegMeta(rr io.Reader) (Meta, error) {
	var tmp [512]byte

	r := asJpegReader(rr)

	if _, err := io.ReadFull(r, tmp[:2]); err != nil {
		return nil, err
	}
	if tmp[0] != 0xff || tmp[1] != jpegSoiMarker {
		return nil, JpegFormatError("missing SOI marker")
	}

	for {
		_, err := io.ReadFull(r, tmp[:2])
		if err != nil {
			return nil, err
		}

		// This is not a segment, continue searching
		for tmp[0] != 0xff {
			tmp[0] = tmp[1]
			tmp[1], err = r.ReadByte()
			if err != nil {
				return nil, err
			}
		}

		marker := tmp[1]

		if marker == 0 {
			// Treat "\xff\x00" as extraneous data.
			continue
		}

		// Marker can be preceded by fill bytes
		for marker == 0xff {
			marker, err = r.ReadByte()
			if err != nil {
				return nil, err
			}
		}

		if marker == jpegEoiMarker { // End Of Image.
			return nil, JpegFormatError("missing SOF marker")
		}

		if jpegRst0Marker <= marker && marker <= jpegRst7Marker {
			continue
		}

		if _, err = io.ReadFull(r, tmp[:2]); err != nil {
			return nil, err
		}
		n := int(tmp[0])<<8 + int(tmp[1]) - 2
		if n <= 0 {
			// We should fail here, but libvips is more tolerant to this, so, continue
			continue
		}

		switch marker {
		case jpegSof0Marker, jpegSof1Marker, jpegSof2Marker, jpegSof3Marker, jpegSof5Marker,
			jpegSof6Marker, jpegSof7Marker, jpegSof9Marker, jpegSof10Marker, jpegSof11Marker,
			jpegSof13Marker, jpegSof14Marker, jpegSof15Marker:
			if _, err := io.ReadFull(r, tmp[:5]); err != nil {
				return nil, err
			}
			// We only support 8-bit precision.
			if tmp[0] != 8 {
				return nil, JpegFormatError("unsupported precision")
			}

			return &meta{
				format: imagetype.JPEG,
				width:  int(tmp[3])<<8 + int(tmp[4]),
				height: int(tmp[1])<<8 + int(tmp[2]),
			}, nil

		case jpegSosMarker:
			return nil, JpegFormatError("missing SOF marker")
		}

		// Skip any other uninteresting segments
		if _, err := r.Discard(n); err != nil {
			return nil, err
		}
	}
}

func init() {
	RegisterFormat("\xff\xd8", DecodeJpegMeta)
}
