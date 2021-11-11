package imagemeta

import (
	"bufio"
	"io"

	"github.com/imgproxy/imgproxy/v3/imagetype"
)

const (
	jpegSof0Marker = 0xc0 // Start Of Frame (Baseline Sequential).
	jpegSof2Marker = 0xc2 // Start Of Frame (Progressive).
	jpegRst0Marker = 0xd0 // ReSTart (0).
	jpegRst7Marker = 0xd7 // ReSTart (7).
	jpegSoiMarker  = 0xd8 // Start Of Image.
	jpegEoiMarker  = 0xd9 // End Of Image.
	jpegSosMarker  = 0xda // Start Of Scan.
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
			// We should fail here, but libvips if more tolerant to this, so, contunue
			continue
		}

		if marker >= jpegSof0Marker && marker <= jpegSof2Marker {
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
		}

		if marker == jpegSosMarker {
			return nil, JpegFormatError("missing SOF marker")
		}

		if n > 0 {
			if _, err := r.Discard(n); err != nil {
				return nil, err
			}
		}
	}
}

func init() {
	RegisterFormat("\xff\xd8", DecodeJpegMeta)
}
